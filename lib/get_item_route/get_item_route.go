// Supports proxying the GetItem endpoint.
package get_item_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	"github.com/smugmug/godynamo/authreq"
	ep "github.com/smugmug/godynamo/endpoint"
	get "github.com/smugmug/godynamo/endpoints/get_item"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the GetItem request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, get.GETITEM_ENDPOINT)
}

// GetItemHandler relays the GetItem request to Dynamo but first validates it through a local type.
func GetItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "get_item_route.GetItemHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "get_item_route.GetItemHandler:cannot parse path."
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("get_item_route.GetItemHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	g := get.NewGetItem()

	um_err := json.Unmarshal(bodybytes, g)
	if um_err != nil {
		e := fmt.Sprintf("get_item_route.GetItemHandler unmarshal err on %s to Get %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := g.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("get_item_route.GetItemHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "get_item_route.GetItemHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		get.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("get_item_route.GetItemHandler %s",
			mr_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
}

// BBPD-only endpoint.
// GetItemJSONHandler issues a GetItem request to aws and then transforms the Response into
// a ResponseItemJSON.
func GetItemJSONHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := authreq.RetryReqJSON_V4(bodybytes, get.GETITEM_ENDPOINT)

	if resp_err != nil {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler: resp err calling %s err %s (input json: %s)",
			get.GETITEM_ENDPOINT, resp_err.Error(), string(bodybytes))
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler: http err %d calling %s (input json: %s)",
			code, get.GETITEM_ENDPOINT, string(bodybytes))
		route_response.WriteError(w, code, e, resp_body)
		return
	}

	// translate the Response to a ResponseItemJSON
	resp := get.NewResponse()
	um_err := json.Unmarshal([]byte(resp_body), resp)
	if um_err != nil {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler:err %s",
			um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	resp_json, rerr := resp.ToResponseItemJSON()
	if rerr != nil {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler:err %s",
			rerr.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	json_body, jerr := json.Marshal(resp_json)
	if jerr != nil {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler:err %s",
			jerr.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		json_body,
		http.StatusOK,
		start,
		get.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("get_item_route.GetItemJSONHandler %s",
			mr_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
}
