//go:build linux || darwin
// +build linux darwin

package signal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

func ListenSignal(termFunc func(), reloadFunc func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	for sig := range c {
		glog.Infof("capture signal: %v", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			termFunc()
		case syscall.SIGUSR1:
			reloadFunc()
		}
	}
}
