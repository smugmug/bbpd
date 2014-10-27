// Supports proxying the UpdateItem endpoint.
package update_item_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	update_item "github.com/smugmug/godynamo/endpoints/update_item"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the UpdateItem request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, update_item.UPDATEITEM_ENDPOINT)
}

// UpdateItemHandler relays the UpdateItem request to Dynamo but first validates it through a local type.
func UpdateItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "update_item_route.UpdateItemHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "update_item_route.UpdateItemHandler:cannot parse path. try /update-item"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("update_item_route.UpdateItemHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	
	u := update_item.NewUpdateItem()
	um_err := json.Unmarshal(bodybytes, u)

	if um_err != nil {
		e := fmt.Sprintf("update_item_route.UpdateItemHandler unmarshal err on %s to Update: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := u.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("update_item_route.UpdateItemHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "update_item_route.UpdateItemHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		update_item.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("update_item_route.UpdateItemHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
