package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// watcher初始化的时候需要读取配置并返回一次configChannel
type Watcher interface {
	watch(filename string, configChannel chan<- map[string]interface{}) error
}

func watchConfig(filename string, configChannel chan<- map[string]interface{}) error {
	var watcher Watcher
	// 这里可以声明更多的watcher，根据filename选择使用哪种类型的监听器
	watcher = &FileWatcher{}

	return watcher.watch(filename, configChannel)
}

type FileWatcher struct{}

func (f FileWatcher) watch(filename string, configChannel chan<- map[string]interface{}) error {
	glog.Infof("watch %s", filename)
	vp := viper.New()
	vp.SetConfigFile(filename)
	vp.WatchConfig()
	vp.OnConfigChange(func(in fsnotify.Event) {
		configChannel <- vp.AllSettings()
	})

	err := vp.ReadInConfig()
	if err != nil {
		return err
	}

	configChannel <- vp.AllSettings()
	return nil
}
