// Supports proxying the DescribeTable endpoint.
package describe_table_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_msg"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	desc "github.com/smugmug/godynamo/endpoints/describe_table"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RawPostHandler relays the DescribeTable request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, desc.DESCTABLE_ENDPOINT)
}

// StatusTableHandler is not a standard endpoint. It can be used to poll a table for readiness
// after a CreateTable or UpdateTable request.
func StatusTableHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "GET" {
		e := "describe_table_route.StatusTableHandler:method only supports GET"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 3 {
		e := "describe_table_route.StatusTableHandler:cannot parse path. try /status-table/TABLENAME"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	ue_tn, ue_err := url.QueryUnescape(string(pathElts[2]))
	if ue_err != nil {
		e := fmt.Sprintf("cannot unescape %s, %s",
			string(pathElts[2]), ue_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	status := "ACTIVE" // our default key
	if query_status, status_ok := req.URL.Query()["status"]; status_ok {
		status = query_status[0]
	}

	poll := false
	if query_poll, poll_ok := req.URL.Query()["poll"]; poll_ok {
		if query_poll[0] == "1" || query_poll[0] == "yes" {
			poll = true
		}
	}

	tries := 1
	if poll {
		tries = 50
	}
	is_status, status_err := desc.PollTableStatus(
		ue_tn,
		status,
		tries)

	if status_err != nil {
		e := fmt.Sprintf("describe_table_route.StatusTableHandler:cannot get status %s from %s, err %s", status, ue_tn, status_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	s := desc.StatusResult{StatusResult: is_status}
	sj, sjerr := json.Marshal(s)
	if sjerr != nil {
		e := fmt.Sprintf("describe_table_route.StatusTableHandler:cannot get convert status to json, err %s", sjerr.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	end := time.Now()
	duration := fmt.Sprintf("%v",end.Sub(start))
	w.Header().Set(bbpd_const.CONTENTTYPE, bbpd_const.JSONMIME)
	b, json_err := json.Marshal(bbpd_msg.Response{
		Name:       desc.ENDPOINT_NAME,
		StatusCode: http.StatusOK,
		Body:       string(sj),
		Run: bbpd_msg.RunInfo{Method: req.Method,
			Host:     bbpd_const.LOCALHOST,
			Duration: duration,
			Start:    start,
			End:      end}})
	if json_err != nil {
		e := fmt.Sprintf("describe_table_route.StatusTableHandler:desc marshal failure %s", json_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(b))
}

// DescribeTableHandler can be used via POST (passing in JSON) or GET (as /DescribeTable/TableName).
func DescribeTableHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	if req.Method == "POST" {
		describeTable_POST_Handler(w, req)
	} else {
		e := fmt.Sprintf("describe_tables_route.DescribeTablesHandler:bad method %s", req.Method)
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
	}
}

// Executes DescribeTable assuming it were requested with the POST method.
func describeTable_POST_Handler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "describe_table_route.describeTable_POST_Handler:cannot parse path."
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("describe_table_route.describeTable_POST_Handler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	
	d := desc.NewDescribeTable()

	um_err := json.Unmarshal(bodybytes, d)
	if um_err != nil {
		e := fmt.Sprintf("describe_table_route.describeTable_POST_Handler unmarshal err on %s to Get %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := d.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("describe_table_route.describeTable_POST_Handler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "describe_table_route.describeTable_POST_Handler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		desc.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("describe_table_route.describeTable_POST_Handler %s",
			mr_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
}
