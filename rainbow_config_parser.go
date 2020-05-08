package main

import (
	"fmt"
	"git.code.oa.com/rainbow/golang-sdk/confapi"
	"git.code.oa.com/rainbow/golang-sdk/types"
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

type RainbowParser struct{}

func (rp *RainbowParser) parse(filepath string) (map[string]interface{}, error) {
	// 解析七彩石配置
	configFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	fi, _ := configFile.Stat()

	if fi.Size() == 0 {
		return nil, fmt.Errorf("config file (%s) is empty", filepath)
	}

	buffer := make([]byte, fi.Size())
	_, err = configFile.Read(buffer)
	if err != nil {
		return nil, err
	}

	var rainbowConfig RainbowConfig
	defaults.SetDefaults(&rainbowConfig)
	err = yaml.Unmarshal(buffer, &rainbowConfig)
	if err != nil {
		glog.Errorf("parse rainbow config fail, err: %v", err)
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

	config := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(val), &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
