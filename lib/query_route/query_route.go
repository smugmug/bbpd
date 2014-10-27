// Supports proxying the Query endpoint.
package query_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	query "github.com/smugmug/godynamo/endpoints/query"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the Query request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, query.QUERY_ENDPOINT)
}

// QueryHandler relays the Query request to Dynamo but first validates it through a local type.
func QueryHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "query_route.QueryHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "query_route.QueryHandler:cannot parse path. try /create, call as POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("query_route.QueryHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	q := query.NewQuery()
	um_err := json.Unmarshal(bodybytes, q)

	if um_err != nil {
		e := fmt.Sprintf("query_route.QueryHandler unmarshal err on %s to Create: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := q.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("query_route.QueryHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "query_route.QueryHandler", resp_body)
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
		e := fmt.Sprintf("query_route.QueryHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
