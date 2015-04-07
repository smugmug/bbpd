// bbpd_stats will collect information about the running bbpd process
package bbpd_stats

import (
	"fmt"
	"sync"
	"time"
)

type Summary struct {
	StartTime       string
	RunningTime     string
	LongestResponse string
	AverageResponse string
	LastResponse    string
	ResponseCount   string
}

var (
	bbpd_start time.Time

	response_count uint64

	longest_response  uint64
	shortest_response uint64
	average_response  float64

	last_response time.Time

	stat_lock sync.RWMutex
)

func init() {
	shortest_response = 9999999999
	average_response = 0.0
	bbpd_start = time.Now()
}

// AddResponse will add the stat information for a response to the totals.
func AddResponse(start time.Time) {
	duration := time.Since(start)
	duration_ns := uint64(duration.Nanoseconds())

	stat_lock.Lock()
	response_count++
	last_response = time.Now()
	new_average_response := average_response*(float64((response_count-1))/float64(response_count)) +
		float64(duration_ns/response_count)
	average_response = new_average_response
	if duration_ns > longest_response {
		longest_response = duration_ns
	}
	if duration_ns < shortest_response && duration_ns != 0 {
		shortest_response = duration_ns
	}
	stat_lock.Unlock()
}

// GetSummary returns a struct of formatted strings that provide human-readable run stats.
func GetSummary() Summary {
	n := time.Since(bbpd_start)
	stat_lock.RLock()
	longest_response_ms := float64(longest_response) / 1000000
	average_response_ms := float64(average_response) / 1000000
	l := "no requests made yet"
	if response_count > 0 {
		l = fmt.Sprintf("%v, (%v ago)", last_response, time.Since(last_response))
	}
	stat_lock.RUnlock()
	return Summary{
		StartTime:       fmt.Sprintf("%v", bbpd_start),
		RunningTime:     fmt.Sprintf("%s", n.String()),
		LongestResponse: fmt.Sprintf("%.2fms", longest_response_ms),
		AverageResponse: fmt.Sprintf("%.2fms", average_response_ms),
		LastResponse:    fmt.Sprintf("%v", l),
		ResponseCount:   fmt.Sprintf("%d", response_count),
	}
}
