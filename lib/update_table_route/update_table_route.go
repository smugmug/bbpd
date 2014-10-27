// Supports proxying the UpdateItem endpoint.
package update_table_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	update_table "github.com/smugmug/godynamo/endpoints/update_table"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the UpdateTable request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, update_table.UPDATETABLE_ENDPOINT)
}

// UpdateTableHandler relays the UpdateTable request to Dynamo but first validates it through a local type.
func UpdateTableHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "update_table_route.UpdateTableHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "update_table_route.UpdateTableHandler:cannot parse path. try /update-table"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("update_table_route.UpdateTableHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	u := update_table.NewUpdateTable()
	um_err := json.Unmarshal(bodybytes, u)

	if um_err != nil {
		e := fmt.Sprintf("update_table_route.UpdateTableHandler unmarshal err on %s to Update: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := u.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("update_item_route.UpdateTableHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "update_item_route.UpdateTableHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		update_table.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("update_table_route.UpdateTableHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
