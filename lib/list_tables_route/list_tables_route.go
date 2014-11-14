// Supports proxying the ListTables endpoint.
package list_tables_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	list "github.com/smugmug/godynamo/endpoints/list_tables"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_LIMIT = 99
)

// RawPostHandler relays the ListTables request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, list.LISTTABLE_ENDPOINT)
}

// ListTablesHandler relays the ListTables request to Dynamo but first validates it through a local type.
func ListTablesHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	if req.Method == "GET" {
		listTables_GET_Handler(w, req)
	} else if req.Method == "POST" {
		listTables_POST_Handler(w, req)
	} else {
		e := fmt.Sprintf("list_tables_route.ListTablesHandler:bad method %s", req.Method)
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
	}
}

// Executes ListTables assuming it were requested with the POST method.
func listTables_POST_Handler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "list_tables_route.listTables_POST_Handler:cannot parse path. try /batch-get-item"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("list_tables_route.listTables_POST_Handler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	var l list.List

	um_err := json.Unmarshal(bodybytes, &l)
	if um_err != nil {
		e := fmt.Sprintf("list_tables_route.listTables_POST_Handler unmarshal err on %s to Get %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := l.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("list_table_route.ListTable_POST_Handler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "list_table_route.ListTable_POST_Handler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		list.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("list_tables_route.listTables_POST_Handler %s",
			mr_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
}

// Executes ListTables assuming it were requested with the GET method.
func listTables_GET_Handler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "list_table_route.ListTablesHandler:cannot parse path." +
			"try /list?ExclusiveStartTableName=$T&Limit=$L"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	queryMap := make(map[string]string)
	for k, v := range req.URL.Query() {
		queryMap[strings.ToLower(k)] = v[0]
	}

	q_estn, estn_exists := queryMap[strings.ToLower(list.EXCLUSIVE_START_TABLE_NAME)]
	estn := ""
	if estn_exists {
		estn = q_estn
	}
	q_limit, limit_exists := queryMap[strings.ToLower(list.LIMIT)]
	limit := uint64(0)
	if limit_exists {
		limit_conv, conv_err := strconv.ParseUint(q_limit, 10, 64)
		if conv_err != nil {
			e := fmt.Sprintf("list_table_route.listTables_GET_Handler bad limit %s", q_limit)
			log.Printf(e)
		} else {
			limit = limit_conv
			if limit > DEFAULT_LIMIT {
				e := fmt.Sprintf("list_table_route.listTables_GET_Handler: high limit %d", limit_conv)
				log.Printf(e)
				limit = DEFAULT_LIMIT
			}
		}
	}

	l := list.List{
		Limit: limit,
		ExclusiveStartTableName: estn}

	resp_body, code, resp_err := l.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("list_table_route.ListTable_GET_Handler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "list_table_route.ListTable_GET_Handler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		list.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("list_table_route.listTable_GET_Handler %s", mr_err.Error())
		log.Printf(e)
	}
}
