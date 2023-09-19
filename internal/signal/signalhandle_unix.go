//go:build linux || darwin
// +build linux darwin

package signal

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
)

func ListenSignal(termFunc func(), reloadFunc func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	for sig := range c {
		klog.Infof("capture signal: %v", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			termFunc()
		case syscall.SIGUSR1:
			reloadFunc()
		}
	}
}
