package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
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
				glog.Info(event)
				glog.Info(more)
				if !more {
					glog.Info("no more event from config file watcher")
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					reloadFunc()
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					watcher.Add(filename)
					reloadFunc()
				}
			case err, more := <-watcher.Errors:
				if !more {
					glog.Info("no more event from error channel of config file watcher")
					return
				}
				glog.Errorf("error from config file watcher: %v", err)
			}
		}
	}()

	return nil
}
