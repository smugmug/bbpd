// Copyright (c) 2013,2014 SmugMug, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//     * Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
//       copyright notice, this list of conditions and the following
//       disclaimer in the documentation and/or other materials provided
//       with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY SMUGMUG, INC. ``AS IS'' AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
// PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL SMUGMUG, INC. BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE
// GOODS OR SERVICES;LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
// IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
// OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
// ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Control routines for starting and stopping bbpd safely
package bbpd_runinfo

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	accepting  bool
	accept_mut *sync.RWMutex
)

func init() {
	accepting = false
	accept_mut = new(sync.RWMutex)
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
		e := fmt.Sprintf("bbpd no longer accepting connections")
		log.Printf(e)
		http.Error(w, e, http.StatusServiceUnavailable)
	}
	return closed
}
