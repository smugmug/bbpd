// Copyright (c) 2013, SmugMug, Inc. All rights reserved.
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
	"runtime"
	"log"
	"fmt"
	"syscall"
	"os"
	"time"
	"os/signal"
	"log/syslog"
	"github.com/smugmug/bbpd/lib/bbpd_runinfo"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_route"
	"github.com/bradclawsie/slog"
	conf "github.com/smugmug/godynamo/conf"
	conf_file "github.com/smugmug/godynamo/conf_file"
	conf_iam "github.com/smugmug/godynamo/conf_iam"
)

// handle signals. we prefer 1,2,3,15
func sigHandle(c <-chan os.Signal) {
	select {
	case sig := <-c:
		switch sig.(os.Signal) {
		case syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGHUP,
			syscall.SIGINT:
			log.Printf("*** caught signal %v, stop\n", sig)
			stop_err := bbpd_runinfo.StopBBPD()
			if stop_err != nil {
				slog.SLog(syslog.LOG_ERR,"no server running?",true)
			}
			log.Printf("sleeping for 5 seconds\n")
			time.Sleep(time.Duration(5) * time.Second)
			log.Printf("bbpd exit\n")
			os.Exit(0)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	sigchan := make(chan os.Signal, 1)
        signal.Notify(sigchan)
	go sigHandle(sigchan)
	slog.Log_Hello = "bbpd"

	// conf file must be read in before anything else, to initialize permissions etc
        conf_file.Read()
	if conf.Vals.Initialized == false {
		panic("the conf.Vals global conf struct has not been initialized, " +
			"invoke with conf_file.Read()")
	} else {
		slog.SLog(syslog.LOG_NOTICE,"global conf.Vals initialized",true)
	}

	// the naive "fire and forget" IAM roles initializer and watcher.
	iam_ready_chan := make(chan bool)
	go conf_iam.GoIAM(iam_ready_chan)
	iam_ready := <- iam_ready_chan
	if !iam_ready {
		slog.SLog(syslog.LOG_NOTICE,"IAM not enabled, using access/secret",true)
	}

	slog.SLog(syslog.LOG_NOTICE,"starting bbpd...",true)
	pid := syscall.Getpid()
	e := fmt.Sprintf("stop with ctrl-c or kill -[1,2,3,15] %v",pid)
	slog.SLog(syslog.LOG_NOTICE,e,true)
	ports := []int{bbpd_const.PORT,bbpd_const.PORT2}
	log.Fatal(bbpd_route.StartBBPD(ports))
}
