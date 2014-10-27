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
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("raw_post_route.RawPostReq err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

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
