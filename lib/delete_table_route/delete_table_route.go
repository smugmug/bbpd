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

// Supports proxying the DeleteTable endpoint.
// It is recommended that you do NOT bind this to a route, its too dangerous.
package delete_table_route

import (
	"net/http"
	"net/url"
	"fmt"
	"log/syslog"
	"strings"
	"time"
	"io"
	"io/ioutil"
	"encoding/json"
	"github.com/smugmug/bbpd/lib/route_response"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	delete_table "github.com/smugmug/godynamo/endpoints/delete_table"
	ep "github.com/smugmug/godynamo/endpoint"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/bradclawsie/slog"
)

// RawPostHandler relays the DeleteTable request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w,req,delete_table.DELETETABLE_ENDPOINT)
}

// DeleteTableHandler can be used via POST (passing in JSON) or GET (as /DeleteTable/TableName).
func DeleteTableHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	if (req.Method == "GET") {
		deleteTable_GET_Handler(w,req)
	} else if (req.Method == "POST") {
		deleteTable_POST_Handler(w,req)
	} else {
		e := fmt.Sprintf("delete_table_route.DeleteTablesHandler:bad method %s",req.Method)
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
	}
}

// Executes DeleteTable assuming it were requested with the POST method.
func deleteTable_POST_Handler(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
	pathElts := strings.Split(req.URL.Path,"/")
	if len(pathElts) != 2 {
		e := "delete_table_route.deleteTable_POST_Handler:cannot parse path. try /desc-table"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler err reading req body: %s",read_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
        req.Body.Close()

	var d delete_table.Delete

	um_err := json.Unmarshal(bodybytes, &d)
	if um_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler unmarshal err on %s to Get %s",string(bodybytes),um_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	d_ep := ep.Endpoint(d)
	resp_body,code,resp_err := d_ep.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler:err %s",
			resp_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w, e, http.StatusInternalServerError)
	 	return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w,code,"delete_table_route.deleteTable_POST_Handler",resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
	 	resp_body,
	 	code,
	 	start,
	 	delete_table.ENDPOINT_NAME)
	if mr_err != nil {
	 	e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler %s",
			mr_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	  	http.Error(w, e, http.StatusInternalServerError)
	  	return
	}
}

// Executes DeleteTable assuming it were requested with the GET method (i.e. /DeleteTable/TableName).
func deleteTable_GET_Handler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	pathElts := strings.Split(req.URL.Path,"/")
	if len(pathElts) != 3 {
		e := "delete_table_route.deleteTable_GET_Handler:cannot parse path. try /delete/TABLENAME"
		slog.SLog(syslog.LOG_ERR,e,false)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	ue_tn,ue_err := url.QueryUnescape(string(pathElts[2]))
	if ue_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_GET_Handler:cannot unescape %s, %s",string(pathElts[2]), ue_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	d := ep.Endpoint(delete_table.Delete{TableName:ue_tn})
	resp_body,code,resp_err := d.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_GET_Handler:err %s",
			resp_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w, e, http.StatusInternalServerError)
	 	return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w,code,"delete_table_route.deleteTable_GET_Handler",resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		delete_table.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_GET_Handler %s",mr_err.Error())
		slog.SLog(syslog.LOG_ERR,e,true)
	}
}
