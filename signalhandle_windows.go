// +build windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

func listenSignal() {
	c := make(chan os.Signal, 1)
	var stop bool
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	defer glog.Infof("listen signal stop, exit...")

	for sig := range c {
		glog.Infof("capture signal: %v", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			inputs.stop()
			close(configChannel)
			stop = true
		}

		if stop {
			break
		}
	}
}
