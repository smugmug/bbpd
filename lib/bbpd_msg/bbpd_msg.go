// Some core types for managing proxied requests.
package bbpd_msg

import (
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
	Body       string
	Run        RunInfo
}
