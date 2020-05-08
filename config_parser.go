package main

import (
	"errors"
	"strings"
)

type Config map[string]interface{}

type Parser interface {
	parse(filename string) (map[string]interface{}, error)
}

func parseConfig(filename string) (map[string]interface{}, error) {
	lowerFilename := strings.ToLower(filename)
	if strings.HasSuffix(lowerFilename, ".yaml") || strings.HasSuffix(lowerFilename, ".yml") {
		yp := &YamlParser{}
		return yp.parse(filename)
	} else if strings.HasSuffix(lowerFilename, ".rainbow") {
		rp := &RainbowParser{}
		return rp.parse(filename)
	}

	return nil, errors.New("unknown config format")
}
