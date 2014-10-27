// Supports proxying the DeleteTable endpoint.
// It is recommended that you do NOT bind this to a route, its too dangerous.
package delete_table_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	delete_table "github.com/smugmug/godynamo/endpoints/delete_table"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the DeleteTable request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, delete_table.DELETETABLE_ENDPOINT)
}

// DeleteTableHandler can be used via POST (passing in JSON) or GET (as /DeleteTable/TableName).
func DeleteTableHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	if req.Method == "POST" {
		deleteTable_POST_Handler(w, req)
	} else {
		e := fmt.Sprintf("delete_table_route.DeleteTablesHandler:bad method %s", req.Method)
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
	}
}

// Executes DeleteTable assuming it were requested with the POST method.
func deleteTable_POST_Handler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "delete_table_route.deleteTable_POST_Handler:cannot parse path. try /desc-table"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	d := delete_table.NewDeleteTable()

	um_err := json.Unmarshal(bodybytes, d)
	if um_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler unmarshal err on %s to Get %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := d.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("delete_table_route.deleteTable_POST_Handler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "delete_table_route.deleteTable_POST_Handler", resp_body)
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
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
}
