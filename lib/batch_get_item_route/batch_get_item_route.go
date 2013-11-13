// Copyright (c) 2013, SmugMug, Inc. All rights reserved.
// 
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//     * Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
//       copyright notice, this list of conditions and the following
//       disclaimer in the documentation and/or other materials provided
//       with the distribution.
// 
// THIS SOFTWARE IS PROVIDED BY SMUGMUG, INC. ``AS IS'' AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
// PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL SMUGMUG, INC. BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE
// GOODS OR SERVICES;LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
// IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
// OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
// ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Supports proxying the BatchGetItem endpoint.
package batch_get_item_route

import (
	"net/http"
	"fmt"
	"log/syslog"
	"strings"
	"time"
	"io"
	"io/ioutil"
	"encoding/json"
	"github.com/smugmug/bbpd/lib/route_response"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	bgi "github.com/smugmug/godynamo/endpoints/batch_get_item"
	ep "github.com/smugmug/godynamo/endpoint"
	"github.com/bradclawsie/slog"
)

// BatchGetItemHandler accepts arbitrarily-sized BatchGetItem requests and relays them to Dynamo.
func BatchGetItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if (req.Method != "POST") {
		e := "batch_get_item_route.BatchGetItemHandler:method only supports POST"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path,"/")
	if len(pathElts) != 2 {
		e := "batch_get_item_route.BatchGetItemHandler:cannot parse path. try /batch-get-item"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler err reading req body: %s",read_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
        req.Body.Close()

	var b bgi.BatchGetItem

	um_err := json.Unmarshal(bodybytes, &b)
 	if um_err != nil {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler unmarshal err on %s to BatchGetItem %s",string(bodybytes),um_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if len(bodybytes) > bgi.QUERY_LIM_BYTES {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler - payload over 1024kb, may be rejected by aws! splitting into segmented requests will likely mean each segment is accepted")
		slog.SLog(syslog.LOG_NOTICE,e,true)
	}

	resp_body,code,resp_err := b.DoBatchGet()

	if resp_err != nil {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler:err %s",
			resp_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w, e, http.StatusInternalServerError)
	 	return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w,code,"batch_get_item_route.BatchGetItemHandler",resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		bgi.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler %s",mr_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	}
}
