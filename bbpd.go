// bbpd is a proxy daemon for Amazon's DynamoDB. See ../../README.md
package main

import (
	"fmt"
	"github.com/smugmug/bbpd/lib/bbpd_const"
	"github.com/smugmug/bbpd/lib/bbpd_route"
	conf "github.com/smugmug/godynamo/conf"
	conf_file "github.com/smugmug/godynamo/conf_file"
	conf_iam "github.com/smugmug/godynamo/conf_iam"
	keepalive "github.com/smugmug/godynamo/keepalive"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// handle signals. we prefer 1,3,15 and will panic on 2
func sigHandle(c <-chan os.Signal) {
	select {
	case sig := <-c:
		switch sig.(os.Signal) {
		case syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGHUP:
			log.Printf("*** caught signal %v, stop\n", sig)
			stop_err := bbpd_route.StopBBPD()
			if stop_err != nil {
				log.Printf("graceful shutdown not possible:%s",stop_err.Error())
			}
			log.Printf("bbpd exit\n")
			os.Exit(0)
		case syscall.SIGINT:
			log.Printf("*** caught signal %v, PANIC stop\n", sig)
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
	conf.Vals.ConfLock.RLock()
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

	// we must give up the lock on the conf before calling GoIAM below, or it
	// will not be able to mutate the auth params
	using_iam := (conf.Vals.UseIAM == true)
	conf.Vals.ConfLock.RUnlock()

	// the naive "fire and forget" IAM roles initializer and watcher.
	if using_iam {
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
	start_bbpd_err := bbpd_route.StartBBPD(ports)
	if start_bbpd_err == nil {
		// all ports are in use. exit with 0 so our rc system does not 
		// respawn the program
		log.Printf("all bbpd ports appear to be in use: exit with code 0")
		os.Exit(0)
	} else {
		// abnormal exit - allow the rc system to try to respawn by returning 
		// exit code 1
		log.Printf("bbpd invocation error")
		log.Fatal(start_bbpd_err.Error())
	}
}
