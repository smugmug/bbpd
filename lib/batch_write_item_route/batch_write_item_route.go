// Supports proxying the BatchWriteItem endpoint.
package batch_write_item_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	bwi "github.com/smugmug/godynamo/endpoints/batch_write_item"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// BatchWriteItemHandler accepts arbitrarily-sized BatchWriteItem requests and relays them to Dynamo.
func BatchWriteItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "batch_write_item_route.BatchWriteItemHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "batch_write_item_route.BatchWriteItemHandler:cannot parse path. try /batch-get-item"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	req.Body.Close()

	if len(bodybytes) > bwi.QUERY_LIM_BYTES {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler - payload over 1024kb, may be rejected by aws! splitting into segmented requests will likely mean each segment is accepted")
		log.Printf(e)
	}

	var b bwi.BatchWriteItem

	um_err := json.Unmarshal(bodybytes, &b)
	if um_err != nil {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler unmarshal err on %s to BatchWriteItem %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := b.DoBatchWrite()

	if resp_err != nil {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "batch_write_item_route.BatchWriteItemHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		bwi.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("batch_write_item_route.BatchWriteItemHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
