// Supports proxying the PutItem endpoint.
package put_item_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the PutItem request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, put.PUTITEM_ENDPOINT)
}

// PutItemHandler relays the PutItem request to Dynamo but first validates it through a local type.
func PutItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "put_item_route.PutItemHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "put_item_route.PutItemHandler:cannot parse path. try /put-item, call as POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("put_item_route.PutItemHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	p := put.NewPutItem()
	um_err := json.Unmarshal(bodybytes, p)

	if um_err != nil {
		e := fmt.Sprintf("put_item_route.PutItemHandler unmarshal err on %s to PutExpected: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := p.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("put_item_route.PutItemHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "put_item_route.PutItemHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		put.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("put_item_route.PutItemHandler %s", mr_err.Error())
		log.Printf(e)
	}
}

// BBPD-only endpoint.
// PutItemJSONHandler relays the PutItem request to Dynamo but first validates it through a local type.
// This variant allows the Item to be encoded as basic JSON. As there is always a conversion that
// needs to be performed from a PutIemJSON struct to a PutItem, this endpoint cannot utilize
// RawPost.
func PutItemJSONHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "put_item_route.PutItemJSONHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "put_item_route.PutItemJSONHandler:cannot parse path. try /put-item, call as POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("put_item_route.PutItemJSONHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	p_json := put.NewPutItemJSON()
	um_err := json.Unmarshal(bodybytes, p_json)

	if um_err != nil {
		e := fmt.Sprintf("put_item_route.PutItemJSONHandler unmarshal err on %s to PutExpected: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	p, perr := p_json.ToPutItem()
	if perr != nil {
		e := fmt.Sprintf("put_item_route.PutItemJSONHandler cannot convert PutItemJSON to PutItem:%s", perr.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := p.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("put_item_route.PutItemJSONHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "put_item_route.PutItemJSONHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		put.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("put_item_route.PutItemJSONHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
