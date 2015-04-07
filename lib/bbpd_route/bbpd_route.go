// The core configuration of the http proxy.
package bbpd_route

import (
	"encoding/json"
	"fmt"
	"github.com/smugmug/bbpd/lib/batch_get_item_route"
	"github.com/smugmug/bbpd/lib/batch_write_item_route"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	"github.com/smugmug/bbpd/lib/bbpd_stats"
	"github.com/smugmug/bbpd/lib/create_table_route"
	"github.com/smugmug/bbpd/lib/delete_item_route"
	"github.com/smugmug/bbpd/lib/describe_table_route"
	"github.com/smugmug/bbpd/lib/get_item_route"
	"github.com/smugmug/bbpd/lib/list_tables_route"
	"github.com/smugmug/bbpd/lib/put_item_route"
	"github.com/smugmug/bbpd/lib/query_route"
	"github.com/smugmug/bbpd/lib/raw_post_route"
	"github.com/smugmug/bbpd/lib/route_response"
	"github.com/smugmug/bbpd/lib/scan_route"
	"github.com/smugmug/bbpd/lib/update_item_route"
	"github.com/smugmug/bbpd/lib/update_table_route"
	"github.com/smugmug/godynamo/aws_const"
	bgi "github.com/smugmug/godynamo/endpoints/batch_get_item"
	bwi "github.com/smugmug/godynamo/endpoints/batch_write_item"
	create "github.com/smugmug/godynamo/endpoints/create_table"
	delete_item "github.com/smugmug/godynamo/endpoints/delete_item"
	desc "github.com/smugmug/godynamo/endpoints/describe_table"
	get "github.com/smugmug/godynamo/endpoints/get_item"
	list "github.com/smugmug/godynamo/endpoints/list_tables"
	put "github.com/smugmug/godynamo/endpoints/put_item"
	query "github.com/smugmug/godynamo/endpoints/query"
	scan "github.com/smugmug/godynamo/endpoints/scan"
	update_item "github.com/smugmug/godynamo/endpoints/update_item"
	update_table "github.com/smugmug/godynamo/endpoints/update_table"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	delete_table "github.com/smugmug/godynamo/endpoints/delete_table"
	// undelete to enable table deletions (dangerous!)
	// "github.com/smugmug/bbpd/lib/delete_table_route"
)

const (
	URI_PATH_SEP           = "/"
	STATUSPATH             = URI_PATH_SEP + "Status"
	STATUSTABLEPATH        = URI_PATH_SEP + "StatusTable" + URI_PATH_SEP
	RAWPOSTPATH            = URI_PATH_SEP + "RawPost" + URI_PATH_SEP
	DESCRIBETABLEPATH      = URI_PATH_SEP + desc.ENDPOINT_NAME
	DESCRIBETABLEGETPATH   = URI_PATH_SEP + desc.ENDPOINT_NAME + URI_PATH_SEP
	DELETETABLEPATH        = URI_PATH_SEP + delete_table.ENDPOINT_NAME
	DELETETABLEGETPATH     = URI_PATH_SEP + delete_table.ENDPOINT_NAME + URI_PATH_SEP
	LISTTABLESPATH         = URI_PATH_SEP + list.ENDPOINT_NAME
	CREATETABLEPATH        = URI_PATH_SEP + create.ENDPOINT_NAME
	UPDATETABLEPATH        = URI_PATH_SEP + update_table.ENDPOINT_NAME
	PUTITEMPATH            = URI_PATH_SEP + put.ENDPOINT_NAME
	PUTITEMJSONPATH        = URI_PATH_SEP + put.JSON_ENDPOINT_NAME
	GETITEMPATH            = URI_PATH_SEP + get.ENDPOINT_NAME
	GETITEMJSONPATH        = URI_PATH_SEP + get.JSON_ENDPOINT_NAME
	BATCHGETITEMPATH       = URI_PATH_SEP + bgi.ENDPOINT_NAME
	BATCHGETITEMJSONPATH   = URI_PATH_SEP + bgi.JSON_ENDPOINT_NAME
	BATCHWRITEITEMPATH     = URI_PATH_SEP + bwi.ENDPOINT_NAME
	BATCHWRITEITEMJSONPATH = URI_PATH_SEP + bwi.JSON_ENDPOINT_NAME
	DELETEITEMPATH         = URI_PATH_SEP + delete_item.ENDPOINT_NAME
	UPDATEITEMPATH         = URI_PATH_SEP + update_item.ENDPOINT_NAME
	QUERYPATH              = URI_PATH_SEP + query.ENDPOINT_NAME
	SCANPATH               = URI_PATH_SEP + scan.ENDPOINT_NAME
	COMPATPATH             = URI_PATH_SEP
)

var (
	availableGetHandlers  []string
	availablePostHandlers []string
	availableHandlers     []string
	srv                   *http.Server
	port                  *int
)

type Status_Struct struct {
	Status            string
	AvailableHandlers []string
	Args              map[string]string
	Summary           bbpd_stats.Summary
}

func init() {
	// we want this to be initialized to be unuseable
	port = nil
	// available handlers
	availableGetHandlers = []string{
		DESCRIBETABLEGETPATH,
	}
	availablePostHandlers = []string{
		DELETEITEMPATH,
		LISTTABLESPATH,
		CREATETABLEPATH,
		UPDATETABLEPATH,
		STATUSTABLEPATH,
		PUTITEMPATH,
		PUTITEMJSONPATH,
		GETITEMPATH,
		GETITEMJSONPATH,
		BATCHGETITEMPATH,
		BATCHGETITEMJSONPATH,
		BATCHWRITEITEMPATH,
		BATCHWRITEITEMJSONPATH,
		UPDATEITEMPATH,
		QUERYPATH,
		SCANPATH,
		RAWPOSTPATH,
		COMPATPATH,
	}
	availableHandlers = append(availableHandlers, availableGetHandlers...)
	availableHandlers = append(availableHandlers, availablePostHandlers...)
	srv = nil
}

func Get_port() int {
	if port == nil {
		return 0
	} else {
		return *port
	}
}

// StatusHandler displays available handlers.
func statusHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	if req.Method != "GET" {
		e := "method only supports GET"
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	var ss Status_Struct
	ss.Args = make(map[string]string)
	ss.Status = "ready"

	ss.Args[bbpd_const.X_BBPD_VERBOSE] = "set '-H \"X-Bbpd-Verbose: True\" ' to get verbose output"
	ss.Args[bbpd_const.X_BBPD_INDENT] = "set '-H \"X-Bbpd-Indent: True\" ' to indent the top-level json"
	ss.AvailableHandlers = availableHandlers
	ss.Summary = bbpd_stats.GetSummary()
	sj, sj_err := json.Marshal(ss)
	if sj_err != nil {
		e := fmt.Sprintf("bbpd_route.statusHandler:status marshal err %s", sj_err.Error())
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	mr_err := route_response.MakeRouteResponse(
		w,
		req,
		sj,
		http.StatusOK,
		time.Now(),
		"Status")
	if mr_err != nil {
		e := fmt.Sprintf("bbpd_route.StatusHandler %s", mr_err.Error())
		log.Printf(e)
	}
}

// can we use this port?
func canAssignPort(requestedPort int) bool {
	_, err := net.Dial("tcp", bbpd_const.LOCALHOST+":"+strconv.Itoa(requestedPort))
	return err != nil
}

// CompatHandler allows bbpd to act as a partial pass-through proxy. Users can provide
// their own body and endpoint target header, but other headers are ignored.
// To use this, set headers with your http client. For example, with curl:
// curl -H "X-Amz-Target: DynamoDB_20120810.DescribeTable" -X POST -d '{"TableName":"mytable"}' http://localhost:12333/
// or alternately
// curl -H "X-Amz-Target: DescribeTable" -X POST -d '{"TableName":"mytable"}' http://localhost:12333/
// if you wish to just use the default API version string.
func CompatHandler(w http.ResponseWriter, req *http.Request) {
	if bbpd_runinfo.BBPDAbortIfClosed(w) {
		return
	}
	// look for the X-Amz-Target header
	target_, target_ok := req.Header[aws_const.AMZ_TARGET_HDR]
	if !target_ok {
		e := fmt.Sprintf("bbpd_route.CompatHandler:missing %s", aws_const.AMZ_TARGET_HDR)
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	target := target_[0]
	normalized_target := target
	target_version_delim := "."
	// allow the header to have the API version string or not
	if strings.Contains(target, target_version_delim) {
		vers_target := strings.SplitN(target, target_version_delim, 2)
		if vers_target[0] != aws_const.CURRENT_API_VERSION {
			e := fmt.Sprintf("bbpd_route.CompatHandler:unsupported API version '%s'", vers_target[0])
			log.Printf(e)
			http.Error(w, e, http.StatusBadRequest)
			return
		}
		normalized_target = vers_target[1]
	}
	endpoint_path := "/" + normalized_target
	if endpoint_path == COMPATPATH || normalized_target == "" {
		e := fmt.Sprintf("bbpd_route.CompatHandler:must call named endpoint")
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}

	// call the proper handler for the header
	switch endpoint_path {
	case DESCRIBETABLEPATH:
		describe_table_route.RawPostHandler(w, req)
		return
	case LISTTABLESPATH:
		list_tables_route.RawPostHandler(w, req)
		return
	case CREATETABLEPATH:
		create_table_route.RawPostHandler(w, req)
		return
	case UPDATETABLEPATH:
		update_table_route.RawPostHandler(w, req)
		return
	case STATUSTABLEPATH:
		describe_table_route.StatusTableHandler(w, req)
		return
	case PUTITEMPATH:
		put_item_route.RawPostHandler(w, req)
		return
	case GETITEMPATH:
		get_item_route.RawPostHandler(w, req)
		return
	case BATCHGETITEMPATH:
		batch_get_item_route.BatchGetItemHandler(w, req)
		return
	case BATCHWRITEITEMPATH:
		batch_write_item_route.BatchWriteItemHandler(w, req)
		return
	case DELETEITEMPATH:
		delete_item_route.RawPostHandler(w, req)
		return
	case UPDATEITEMPATH:
		update_item_route.RawPostHandler(w, req)
		return
	case QUERYPATH:
		query_route.RawPostHandler(w, req)
		return
	case SCANPATH:
		scan_route.RawPostHandler(w, req)
		return
	default:
		e := fmt.Sprintf("bbpd_route.CompatHandler:unknown endpoint '%s'", endpoint_path)
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return
	}
}

// StartBBPD is where the proxy http server is started.
// The requestedPort is a *int so it can be nil'able. passing 0 as an implied
// null value could result in a dial that takes any available port, as is
// implied by the go docs.
func StartBBPD(requestedPorts []int) error {
	// try to get a port to listen to
	for _, p := range requestedPorts {
		e := fmt.Sprintf("trying to bind to port:%d", p)
		log.Printf(e)
		if canAssignPort(p) {
			port = &p
			break
		} else {
			e := fmt.Sprintf("port %d already in use", p)
			log.Printf(e)
		}
	}
	if port == nil {
		// if all ports are in use, we may assume that other bbpd invocations are
		// running correctly. in which case, return nil here and the caller will
		// exit with code 0, which is important to prevent rc managers etc from
		// automatically respawning the program
		log.Printf("bbpd_route.StartBBPD:no listen port")
		return nil
	}
	e := fmt.Sprintf("init routing on port %d", *port)
	log.Printf(e)
	http.HandleFunc(STATUSPATH, statusHandler)
	http.HandleFunc(DESCRIBETABLEPATH, describe_table_route.RawPostHandler)
	http.HandleFunc(DESCRIBETABLEGETPATH, describe_table_route.DescribeTableHandler)
	http.HandleFunc(LISTTABLESPATH, list_tables_route.ListTablesHandler)
	http.HandleFunc(CREATETABLEPATH, create_table_route.RawPostHandler)
	http.HandleFunc(UPDATETABLEPATH, update_table_route.RawPostHandler)
	http.HandleFunc(STATUSTABLEPATH, describe_table_route.StatusTableHandler)
	http.HandleFunc(PUTITEMPATH, put_item_route.RawPostHandler)
	http.HandleFunc(PUTITEMJSONPATH, put_item_route.PutItemJSONHandler)
	http.HandleFunc(GETITEMPATH, get_item_route.RawPostHandler)
	http.HandleFunc(GETITEMJSONPATH, get_item_route.GetItemJSONHandler)
	http.HandleFunc(BATCHGETITEMPATH, batch_get_item_route.BatchGetItemHandler)
	http.HandleFunc(BATCHGETITEMJSONPATH, batch_get_item_route.BatchGetItemJSONHandler)
	http.HandleFunc(BATCHWRITEITEMPATH, batch_write_item_route.BatchWriteItemHandler)
	http.HandleFunc(BATCHWRITEITEMJSONPATH, batch_write_item_route.BatchWriteItemJSONHandler)
	http.HandleFunc(DELETEITEMPATH, delete_item_route.RawPostHandler)
	http.HandleFunc(UPDATEITEMPATH, update_item_route.RawPostHandler)
	http.HandleFunc(QUERYPATH, query_route.RawPostHandler)
	http.HandleFunc(SCANPATH, scan_route.RawPostHandler)
	http.HandleFunc(RAWPOSTPATH, raw_post_route.RawPostHandler)
	http.HandleFunc(COMPATPATH, CompatHandler)

	// undelete these to enable table deletions, a little dangerous!
	// http.HandleFunc(DELETETABLEPATH,     delete_table_route.RawPostHandler)
	// http.HandleFunc(DELETETABLEGETPATH,  delete_table_route.DeleteTableHandler)

	const SERV_TIMEOUT = 20
	srv = &http.Server{
		Addr: ":" + strconv.Itoa(*port),
		// The timeouts seems too-long, but they accomodates the exponential decay retry loop.
		// Programs using this can either change these directly or use goroutine timeouts
		// to impose a local minimum.
		ReadTimeout:  SERV_TIMEOUT * time.Second,
		WriteTimeout: SERV_TIMEOUT * time.Second,
		ConnState: func(conn net.Conn, new_state http.ConnState) {
			bbpd_runinfo.RecordConnState(new_state)
			return
		},
	}
	bbpd_runinfo.SetBBPDAccept()
	return srv.ListenAndServe()
}

func StopBBPD() error {
	return bbpd_runinfo.StopBBPD()
}
