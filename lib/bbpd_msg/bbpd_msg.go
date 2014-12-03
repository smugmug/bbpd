// Some core types for managing proxied requests.
package bbpd_msg

import (
	"encoding/json"
	"time"
)

// RunInfo provides duration information.
type RunInfo struct {
	Method   string
	Host     string
	Start    time.Time
	End      time.Time
	Duration string
}

type Status struct {
	Status string
	Run    RunInfo
}

type Response struct {
	Name       string
	StatusCode int
	Body       []byte
	Run        RunInfo
}

type response struct {
	Name       string
	StatusCode int
	Body       string
	Run        RunInfo
}

// Body is a []byte above so it needs to be converted here or it will be encoded as a bytestream
func (r Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(response{Name: r.Name, StatusCode: r.StatusCode, Run: r.Run, Body: string(r.Body)})
}
