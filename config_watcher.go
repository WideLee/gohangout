package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"strings"
)

// watcher初始化的时候需要读取配置并返回一次configChannel
type Watcher interface {
	watch(filename string, configChannel chan<- map[string]interface{}) error
}

func watchConfig(filename string, configChannel chan<- map[string]interface{}) error {
	lowerFilename := strings.ToLower(filename)
	var watcher Watcher

	if strings.HasSuffix(lowerFilename, ".rainbow") {
		watcher = &Rainbow{}
	} else {
		watcher = &FileWatcher{}
	}

	return watcher.watch(filename, configChannel)
}

type FileWatcher struct{}

func (f FileWatcher) watch(filename string, configChannel chan<- map[string]interface{}) error {
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
