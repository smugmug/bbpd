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

// Supports proxying the BatchWriteItem endpoint.
package batch_write_item_route

import (
	"net/http"
	"fmt"
	"log/syslog"
	"strings"
	"time"
	"io"
	"io/ioutil"
	"encoding/json"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	"github.com/smugmug/bbpd/lib/route_response"
	bwi "github.com/smugmug/godynamo/endpoints/batch_write_item"
	ep "github.com/smugmug/godynamo/endpoint"
	"github.com/bradclawsie/slog"
)

// BatchWriteItemHandler accepts arbitrarily-sized BatchWriteItem requests and relays them to Dynamo.
func BatchWriteItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if (req.Method != "POST") {
		e := "batch_write_item_route.BatchWriteItemHandler:method only supports POST"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path,"/")
	if len(pathElts) != 2 {
		e := "batch_write_item_route.BatchWriteItemHandler:cannot parse path. try /batch-get-item"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler err reading req body: %s",read_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
        req.Body.Close()

	if len(bodybytes) > bwi.QUERY_LIM_BYTES {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler - payload over 1024kb, may be rejected by aws! splitting into segmented requests will likely mean each segment is accepted")
		slog.SLog(syslog.LOG_NOTICE,e,true)
	}

	var b bwi.BatchWriteItem

	um_err := json.Unmarshal(bodybytes, &b)
	if um_err != nil {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler unmarshal err on %s to BatchWriteItem %s",string(bodybytes),um_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body,code,resp_err := b.DoBatchWrite()

	if resp_err != nil {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler:err %s",
			resp_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w, e, http.StatusInternalServerError)
	 	return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w,code,"batch_write_item_route.BatchWriteItemHandler",resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		bwi.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler %s",mr_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	}
}
