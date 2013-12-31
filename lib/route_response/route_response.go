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

package route_response

import (
	"net/http"
	"fmt"
	"log/syslog"
	"io"
	"time"
	"bytes"
	"errors"
	"strconv"
	"encoding/json"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_msg"
	ep "github.com/smugmug/godynamo/endpoint"
	"github.com/bradclawsie/slog"
)

// WriteError is a convenience wrapper for emitting an error.
func WriteError(w http.ResponseWriter,code int,origin,resp_body string) {
	if ep.ReqErr(code) { // 4xx err
		e := fmt.Sprintf("%s:(%d) %s",origin,code,resp_body)
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w,e,http.StatusBadRequest)
	} else { // 5xx err
		e := fmt.Sprintf("%s:(%d) Server Error",origin,code)
		slog.SLog(syslog.LOG_ERR,e,true)
	 	http.Error(w,e,http.StatusInternalServerError)
	}
}

// MakeRouteResponse wraps a dynamo response with some debugging information related to http codes and request duration.
func MakeRouteResponse(w http.ResponseWriter,req *http.Request,resp_body string,code int,start time.Time,endpoint_name string) error {
	duration := fmt.Sprintf("%v",time.Now().Sub(start))
	if resp_body != "" && code == http.StatusOK {
		var b []byte;
		var json_err error;
		w.Header().Set(bbpd_const.CONTENTTYPE, bbpd_const.JSONMIME)
		if _,compact := req.URL.Query()[bbpd_const.COMPACT]; compact {
			b = []byte(resp_body)
		} else {
			b, json_err = json.Marshal(bbpd_msg.Response{
				Name:endpoint_name,
				StatusCode:code,
				Body:resp_body,
				Run:bbpd_msg.RunInfo{Method:req.Method,
					Host:bbpd_const.Host,
					Start:start,
					Duration:duration}})
			if json_err != nil {
				e := fmt.Sprintf("route_response.MakeRouteResponse:marshal failure %s",
					json_err.Error())
				slog.SLog(syslog.LOG_ERR,e,true)
				http.Error(w, e, http.StatusInternalServerError)
				return json_err
			}
		}

		// we support pretty-printing (indent)
		// just pass indent=1 (the 1 can be anything) in the url
		if _,indent := req.URL.Query()[bbpd_const.INDENT]; indent {
			var buf bytes.Buffer
			if i_err := json.Indent(&buf,b,"","\t"); i_err != nil {
				// could not pretty print!
				e := fmt.Sprintf("route_response.MakeRouteResponse cannot indent %s",string(b))
				slog.SLog(syslog.LOG_ERR,e,true)
				unindented_str := string(b)
				w.Header().Set(bbpd_const.CONTENTLENGTH,
					strconv.Itoa(len(unindented_str)))
					io.WriteString(w,unindented_str)
			} else {
				// do the pretty print
				indented_str := buf.String()
				w.Header().Set(bbpd_const.CONTENTLENGTH,
					strconv.Itoa(len(indented_str)))
					io.WriteString(w,indented_str)
			}
		} else {
			// no pretty print requested
			unindented_str := string(b)
			w.Header().Set(bbpd_const.CONTENTLENGTH,
				strconv.Itoa(len(unindented_str)))
				io.WriteString(w,unindented_str)
		}
		return nil
	} else {
		s := ""
		if resp_body != "" {
			s = resp_body
		}
		e := fmt.Sprintf("route_response.MakeRouteResponse %s",s)
		slog.SLog(syslog.LOG_ERR,e,true)
		http.Error(w, e, http.StatusBadRequest)
		return errors.New(e)
	}
}
