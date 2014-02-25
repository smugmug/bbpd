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

package raw_post_route

import (
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	"github.com/smugmug/bbpd/lib/route_response"
	"github.com/smugmug/godynamo/authreq"
	ep "github.com/smugmug/godynamo/endpoint"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RawPostHandler relays POST data directly to Dynamo, typically called by other endpoint proxy packages
// that are recognized by the string in the request path.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	if req.Method != "POST" {
		e := fmt.Sprintf("raw_post_route.RawPostHandler: method only supports POST")
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 3 {
		e := "raw_post_route.RawPostHandler:cannot parse path"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	ue_ep, ue_err := url.QueryUnescape(string(pathElts[2]))
	if ue_err != nil {
		e := fmt.Sprintf("raw_table_route.RawPostHandler:cannot unescape %s, %s", string(pathElts[2]), ue_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	RawPostReq(w, req, ue_ep)
}

// RawPostReq obtains the POST payload from the request and forwards it on to the endpoint amzTarget.
func RawPostReq(w http.ResponseWriter, req *http.Request, amzTarget string) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("raw_post_route.RawPostReq err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	req.Body.Close()

	resp_body, code, resp_err := authreq.RetryReqJSON_V4(bodybytes, amzTarget)

	if resp_err != nil {
		e := fmt.Sprintf("raw_post_route.RawPostReq: resp err calling %s err %s (input json: %s)",
			amzTarget, resp_err.Error(), string(bodybytes))
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		e := fmt.Sprintf("raw_post_route.RawPostReq: http err %d calling %s (input json: %s)",
			code, amzTarget, string(bodybytes))
		route_response.WriteError(w, code, e, resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		amzTarget)
	if mr_err != nil {
		e := fmt.Sprintf("raw_post_route.RawPostReq %s", mr_err.Error())
		log.Printf(e)
	}
}
