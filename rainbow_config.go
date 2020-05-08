package main

import (
	"fmt"
	"git.code.oa.com/rainbow/golang-sdk/confapi"
	"git.code.oa.com/rainbow/golang-sdk/config"
	"git.code.oa.com/rainbow/golang-sdk/types"
	"git.code.oa.com/rainbow/golang-sdk/watch"
	"github.com/golang/glog"
	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v2"
	"os"
)

// 七彩石配置中心的配置修改
type RainbowConfig struct {
	AppId       string `yaml:"AppId"`
	UserId      string `yaml:"UserId"`
	UserKey     string `yaml:"UserKey"`
	Group       string `yaml:"Group"`
	RainbowHost string `yaml:"RainbowHost" default:"http://api.rainbow.oa.com:8080"`
	ConfigKey   string `yaml:"ConfigKey" default:"config"`
}

type Rainbow struct{}

func (rp *Rainbow) watch(filename string, configChannel chan<- map[string]interface{}) error {
	rainbowConfig, err := rp.parseRainbowConfig(filename)
	if err != nil {
		return err
	}

	// 去七彩石拉取配置
	rainbow, err := confapi.New(
		types.ConnectStr(rainbowConfig.RainbowHost),
		types.IsUsingLocalCache(true),
		types.IsUsingFileCache(true),

		// 增加签名
		types.OpenSign(true),
		types.AppID(rainbowConfig.AppId),
		types.UserID(rainbowConfig.UserId),
		types.UserKey(rainbowConfig.UserKey),
		types.HmacWay("sha1"),
	)
	if err != nil {
		glog.Errorf("new rainbow api fail, err: %v", err)
		return err
	}

	var watcher = watch.Watcher{
		Key: rainbowConfig.ConfigKey,
		GetOptions: types.GetOptions{
			AppID: rainbowConfig.AppId,
			Group: rainbowConfig.Group,
		},
		CB: func(oldVal watch.Result, newVal []*config.KeyValueItem) error {
			glog.Infof("[watcher] ---------------------\n")
			glog.Infof("[watcher] old value:%+v\n", oldVal)
			glog.Infof("[watcher] new value:")

			var err error
			cfg := make(map[string]interface{})

			for i := 0; i < len(newVal); i++ {
				glog.Infof("%+v", *newVal[i])
				for _, kv := range newVal[i].KeyValues {
					if kv.GetKey() == rainbowConfig.ConfigKey {
						err = yaml.Unmarshal([]byte(kv.GetValue()), &cfg)
						break
					}
				}
			}
			glog.Infof("[watcher] ---------------------\n")

			if err == nil {
				configChannel <- cfg
			} else {
				glog.Errorf("[watcher] read config fail, err: %v", err)
			}

			return err
		},
	}

	if err = rainbow.AddWatcher(watcher); err != nil {
		return err
	}

	return nil
}

func (rp *Rainbow) parseRainbowConfig(filename string) (RainbowConfig, error) {
	var rainbowConfig RainbowConfig

	// 解析七彩石配置
	configFile, err := os.Open(filename)
	if err != nil {
		return rainbowConfig, err
	}
	fi, _ := configFile.Stat()

	if fi.Size() == 0 {
		return rainbowConfig, fmt.Errorf("config file (%s) is empty", filename)
	}

	buffer := make([]byte, fi.Size())
	_, err = configFile.Read(buffer)
	if err != nil {
		return rainbowConfig, err
	}

	defaults.SetDefaults(&rainbowConfig)
	err = yaml.Unmarshal(buffer, &rainbowConfig)
	if err != nil {
		glog.Errorf("parse rainbow config fail, err: %v", err)
		return rainbowConfig, err
	}

	return rainbowConfig, nil
}

func (rp *Rainbow) parse(filename string) (map[string]interface{}, error) {
	rainbowConfig, err := rp.parseRainbowConfig(filename)
	if err != nil {
		return nil, err
	}

	// 去七彩石拉取配置
	rainbow, err := confapi.New(
		types.ConnectStr(rainbowConfig.RainbowHost),
		types.IsUsingLocalCache(true),
		types.IsUsingFileCache(true),

		// 增加签名
		types.OpenSign(true),
		types.AppID(rainbowConfig.AppId),
		types.UserID(rainbowConfig.UserId),
		types.UserKey(rainbowConfig.UserKey),
		types.HmacWay("sha1"),
	)
	if err != nil {
		glog.Errorf("new rainbow api fail, err: %v", err)
		return nil, err
	}

	// 解析七彩石配置
	getOpts := make([]types.AssignGetOption, 0)
	getOpts = append(getOpts, types.WithAppID(rainbowConfig.AppId))
	getOpts = append(getOpts, types.WithGroup(rainbowConfig.Group))
	var val string
	val, err = rainbow.Get(rainbowConfig.ConfigKey, getOpts...)
	if err != nil {
		glog.Errorf("get rainbow config fail, err: %v", err)
		return nil, err
	}

	cfg := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(val), &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
