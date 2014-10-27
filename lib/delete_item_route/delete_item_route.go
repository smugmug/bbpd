// Supports proxying the DeleteItem endpoint.
package delete_item_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	delete_item "github.com/smugmug/godynamo/endpoints/delete_item"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the DeleteItem request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, delete_item.DELETEITEM_ENDPOINT)
}

// DeleteItemHandler relays the DeleteItem request to Dynamo but first validates it through a local type.
func DeleteItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := fmt.Sprintf("delete_item_route.DeleteItemHandler: method only supports POST")
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := fmt.Sprintf("delete_item_route.DeleteItemHandler:cannot parse path. try /delete-item, call as POST")
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("delete_item_route.DeleteItemHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	d := delete_item.NewDelete()
	um_err := json.Unmarshal(bodybytes, d)

	if um_err != nil {
		e := fmt.Sprintf("delete_item_route.DeleteItemHandler unmarshal err on %s to PutExpected: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := d.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("delete_item_route.DeleteItemHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "delete_item_route.DeleteItemHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		delete_item.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("delete_item_route.DeleteItemHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
