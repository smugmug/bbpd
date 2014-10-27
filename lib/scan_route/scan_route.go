// Supports proxying the Scan endpoint.
package scan_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	scan "github.com/smugmug/godynamo/endpoints/scan"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the Scan request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, scan.SCAN_ENDPOINT)
}

// ScanHandler relays the Scan request to Dynamo but first validates it through a local type.
func ScanHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "scan_route.ScanHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "scan_route.ScanHandler:cannot parse path. try /create, call as POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("scan_route.ScanHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	s := scan.NewScan()
	um_err := json.Unmarshal(bodybytes, s)

	if um_err != nil {
		e := fmt.Sprintf("scan_route.ScanHandler unmarshal err on %s to Create: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	resp_body, code, resp_err := s.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("scan_route.ScanHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "scan_route.ScanHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		scan.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("scan_route.ScanHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
