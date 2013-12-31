// Copyright (c) 2013,2014 SmugMug, Inc. All rights reserved.
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

// Supports proxying the Query endpoint.
package query_route

import (
	"net/http"
	"fmt"
	"log/syslog"
	"strings"
	"io"
	"time"
	"io/ioutil"
	"encoding/json"
	"github.com/smugmug/bbpd/lib/route_response"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	query "github.com/smugmug/godynamo/endpoints/query"
	ep "github.com/smugmug/godynamo/endpoint"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/bradclawsie/slog"
)

// RawPostHandler relays the Query request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w,req,query.QUERY_ENDPOINT)
}

// QueryHandler relays the Query request to Dynamo but first validates it through a local type.
func QueryHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if (req.Method != "POST") {
		e := "query_route.QueryHandler:method only supports POST"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path,"/")
	if len(pathElts) != 2 {
		e := "query_route.QueryHandler:cannot parse path. try /create, call as POST"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("query_route.QueryHandler err reading req body: %s",read_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
        req.Body.Close()

	var q query.Query
	um_err := json.Unmarshal(bodybytes,&q)

	if um_err != nil {
		e := fmt.Sprintf("query_route.QueryHandler unmarshal err on %s to Create: %s",string(bodybytes),um_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body,code,resp_err := q.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("query_route.QueryHandler:err %s",
			resp_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w, e, http.StatusInternalServerError)
	 	return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w,code,"query_route.QueryHandler",resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		query.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("query_route.QueryHandler %s",mr_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	}
}
