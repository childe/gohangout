package config

import (
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
)

// Watcher watches the config file and callback f
func WatchConfig(filename string, reloadFunc func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	watcher.Add(filename)

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, more := <-watcher.Events:
				if !more {
					klog.Info("config file watcher closed")
					return
				}
				klog.Infof("capture file watch event: %s", event)
				reloadFunc()

				// filename may be renamed, so add it again
				watcher.Add(filename)
			case err, more := <-watcher.Errors:
				if !more {
					klog.Info("error channel of config file watcher closed")
					return
				}
				klog.Errorf("error from config file watcher: %v", err)
			}
		}
	}()

	return nil
}
