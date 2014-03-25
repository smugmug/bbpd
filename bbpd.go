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

// bbpd is a proxy daemon for Amazon's DynamoDB. See ../../README.md
package main

import (
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_route"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	conf "github.com/smugmug/godynamo/conf"
	conf_file "github.com/smugmug/godynamo/conf_file"
	conf_iam "github.com/smugmug/godynamo/conf_iam"
	keepalive "github.com/smugmug/godynamo/keepalive"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// handle signals. we prefer 1,2,3,15
func sigHandle(c <-chan os.Signal) {
	select {
	case sig := <-c:
		switch sig.(os.Signal) {
		case syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGHUP:
			log.Printf("*** caught signal %v, stop\n", sig)
			stop_err := bbpd_runinfo.StopBBPD()
			if stop_err != nil {
				log.Printf("no server running?")
			}
			log.Printf("sleeping for 5 seconds\n")
			time.Sleep(time.Duration(5) * time.Second)
			log.Printf("bbpd exit\n")
			os.Exit(0)
		case syscall.SIGINT:
			log.Printf("*** caught signal %v, stop\n", sig)
			panic("bbpd panic")
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan)
	go sigHandle(sigchan)

	// conf file must be read in before anything else, to initialize permissions etc
	conf_file.Read()
	if conf.Vals.Initialized == false {
		panic("the conf.Vals global conf struct has not been initialized, " +
			"invoke with conf_file.Read()")
	} else {
		log.Printf("global conf.Vals initialized")
	}

	// launch a background poller to keep conns to aws alive
	if conf.Vals.Network.DynamoDB.KeepAlive {
		log.Printf("launching background keepalive")
		go keepalive.KeepAlive([]string{conf.Vals.Network.DynamoDB.URL})
	}

	// the naive "fire and forget" IAM roles initializer and watcher.
	if conf.Vals.UseIAM {
		iam_ready_chan := make(chan bool)
		go conf_iam.GoIAM(iam_ready_chan)
		iam_ready := <-iam_ready_chan
		if !iam_ready {
			panic("iam is not ready? auth problem")
		}
	} else {
		log.Printf("not using iam, assume credentials hardcoded in conf file")
	}

	log.Printf("starting bbpd...")
	pid := syscall.Getpid()
	e := fmt.Sprintf("induce panic with ctrl-c (kill -2 %v) or graceful termination with kill -[1,3,15] %v",pid,pid)
	log.Printf(e)
	ports := []int{bbpd_const.PORT, bbpd_const.PORT2}
	log.Fatal(bbpd_route.StartBBPD(ports))
}
