// Supports proxying the BatchGetItem endpoint.
package batch_get_item_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	bgi "github.com/smugmug/godynamo/endpoints/batch_get_item"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// BatchGetItemHandler accepts arbitrarily-sized BatchGetItem requests and relays them to Dynamo.
func BatchGetItemHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "batch_get_item_route.BatchGetItemHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "batch_get_item_route.BatchGetItemHandler:cannot parse path. try /batch-get-item"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	req.Body.Close()

	var b bgi.BatchGetItem

	um_err := json.Unmarshal(bodybytes, &b)
	if um_err != nil {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler unmarshal err on %s to BatchGetItem %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if len(bodybytes) > bgi.QUERY_LIM_BYTES {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler - payload over 1024kb, may be rejected by aws! splitting into segmented requests will likely mean each segment is accepted")
		log.Printf(e)
	}

	resp_body, code, resp_err := b.DoBatchGet()

	if resp_err != nil {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "batch_get_item_route.BatchGetItemHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		bgi.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("batch_get_item_route.BatchGetItemHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
