// Control routines for starting and stopping bbpd safely
package bbpd_runinfo

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	accepting  bool
	accept_mut *sync.RWMutex
	conns_wg   *sync.WaitGroup
)

func init() {
	accepting = false
	accept_mut = new(sync.RWMutex)
	conns_wg = new(sync.WaitGroup)
}

// SetBBPDAccept should be called when the server is started.
func SetBBPDAccept() {
	accept_mut.Lock()
	accepting = true
	accept_mut.Unlock()
}

// StopBBPD executes any shutdown tasks.
func StopBBPD() error {
	accept_mut.Lock()
	accepting = false
	accept_mut.Unlock()
	wait_chan := make(chan bool, 1)
	go func() {
		conns_wg.Wait()
		wait_chan <- true
	}()
	select {
	case <-wait_chan:
		log.Printf("conns completed, graceful exit possible")
		return nil
	case <-time.After(1000 * time.Millisecond):
		return errors.New("shutdown timed out")
	}
	return nil
}

// IsAccepting returns the value of server accepting state that can be set when bbpd should
// stop accepting connections.
func IsAccepting() bool {
	accept_mut.RLock()
	a := accepting
	accept_mut.RUnlock()
	return a
}

// BBPDAbortIfClosed returns a 503 if the server accepting state has been set to false.
func BBPDAbortIfClosed(w http.ResponseWriter) bool {
	closed := !IsAccepting()
	if closed {
		e := "bbpd is in a closed state and is no longer accepting connections"
		http.Error(w, e, http.StatusServiceUnavailable)
	}
	return closed
}

// RecordConnState keeps track of the current connection count via a waitgroup.
func RecordConnState(new_state http.ConnState) {
	switch new_state {
	case http.StateNew:
		conns_wg.Add(1)
	case http.StateClosed, http.StateHijacked:
		conns_wg.Done()
	}
	return
}
