package route_response

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_msg"
	ep "github.com/smugmug/godynamo/endpoint"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// WriteError is a convenience wrapper for emitting an error.
func WriteError(w http.ResponseWriter, code int, origin string, resp_body []byte) {
	if ep.ReqErr(code) { // 4xx err
		e := fmt.Sprintf("%s:(%d) %s", origin, code, string(resp_body))
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
	} else { // 5xx err
		e := fmt.Sprintf("%s:(%d) Server Error", origin, code)
		log.Printf(e)
		http.Error(w, e, http.StatusInternalServerError)
	}
}

// MakeRouteResponse wraps a dynamo response with some debugging information related to http codes and request duration.
func MakeRouteResponse(w http.ResponseWriter, req *http.Request, resp_body []byte, code int, start time.Time, endpoint_name string) error {
	end := time.Now()
	duration := fmt.Sprintf("%v", end.Sub(start))
	if resp_body != nil && code == http.StatusOK {
		var b []byte
		var json_err error
		w.Header().Set(bbpd_const.CONTENTTYPE, bbpd_const.JSONMIME)
		if _, compact := req.URL.Query()[bbpd_const.COMPACT]; compact {
			b = resp_body
		} else {
			b, json_err = json.Marshal(bbpd_msg.Response{
				Name:       endpoint_name,
				StatusCode: code,
				Body:       resp_body,
				Run: bbpd_msg.RunInfo{Method: req.Method,
					Host:     bbpd_const.LOCALHOST,
					Duration: duration,
					Start:    start,
					End:      end}})
			if json_err != nil {
				e := fmt.Sprintf("route_response.MakeRouteResponse:marshal failure %s",
					json_err.Error())
				log.Printf(e)
				http.Error(w, e, http.StatusInternalServerError)
				return json_err
			}
		}

		// we support pretty-printing (indent)
		// just pass indent=1 (the 1 can be anything) in the url
		if _, indent := req.URL.Query()[bbpd_const.INDENT]; indent {
			var buf bytes.Buffer
			if i_err := json.Indent(&buf, b, "", "\t"); i_err != nil {
				// could not pretty print!
				e := fmt.Sprintf("route_response.MakeRouteResponse cannot indent %s", string(b))
				log.Printf(e)
				unindented_str := string(b)
				w.Header().Set(bbpd_const.CONTENTLENGTH,
					strconv.Itoa(len(unindented_str)))
				io.WriteString(w, unindented_str)
			} else {
				// do the pretty print
				indented_str := buf.String()
				w.Header().Set(bbpd_const.CONTENTLENGTH,
					strconv.Itoa(len(indented_str)))
				io.WriteString(w, indented_str)
			}
		} else {
			// no pretty print requested
			unindented_str := string(b)
			w.Header().Set(bbpd_const.CONTENTLENGTH,
				strconv.Itoa(len(unindented_str)))
			io.WriteString(w, unindented_str)
		}
		return nil
	} else {
		s := ""
		if resp_body != nil {
			s = string(resp_body)
		}
		e := fmt.Sprintf("route_response.MakeRouteResponse %s", s)
		log.Printf(e)
		http.Error(w, e, http.StatusBadRequest)
		return errors.New(e)
	}
}
