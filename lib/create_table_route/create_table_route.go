// Supports proxying the CreateTable endpoint.
package create_table_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	raw "github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	ep "github.com/smugmug/godynamo/endpoint"
	create "github.com/smugmug/godynamo/endpoints/create_table"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// RawPostHandler relays the CreateTable request to Dynamo directly.
func RawPostHandler(w http.ResponseWriter, req *http.Request) {
	raw.RawPostReq(w, req, create.CREATETABLE_ENDPOINT)
}

// CreateTableHandler relays the CreateTable request to Dynamo but first validates it through a local type.
func CreateTableHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	start := time.Now()
	if req.Method != "POST" {
		e := "create_table_route.CreateTableHandler:method only supports POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	pathElts := strings.Split(req.URL.Path, "/")
	if len(pathElts) != 2 {
		e := "create_table_route.CreateTableHandler:cannot parse path. try /create, call as POST"
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	bodybytes, read_err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if read_err != nil && read_err != io.EOF {
		e := fmt.Sprintf("create_table_route.CreateTableHandler err reading req body: %s", read_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	c := create.NewCreate()
	um_err := json.Unmarshal(bodybytes, c)

	if um_err != nil {
		e := fmt.Sprintf("create_table_route.CreateTableHandler unmarshal err on %s to Create: %s", string(bodybytes), um_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	// the table name can't be too long, 256 bytes binary utf8
	if !create.ValidTableName(c.TableName) {
		e := fmt.Sprintf("create_table_route.CreateTableHandler: tablename over 256 bytes")
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	resp_body, code, resp_err := c.EndpointReq()

	if resp_err != nil {
		e := fmt.Sprintf("create_table_route.CreateTableHandler:err %s",
			resp_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	if ep.HttpErr(code) {
		route_response.WriteError(w, code, "create_table_route.CreateTableHandler", resp_body)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		resp_body,
		code,
		start,
		create.ENDPOINT_NAME)
	if mr_err != nil {
		e := fmt.Sprintf("create_table_route.CreateTableHandler %s", mr_err.Error())
		log.Printf(e)
	}
}
