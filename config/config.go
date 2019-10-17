//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package config is responsible parsing the configuration file,
// and then, it is used by other parts of application.
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Configurator combine every child config
type Configurator struct {
	Database     `yaml:"database,omitempty"`
	Log          `yaml:"log,omitempty"`
	HTTP         `yaml:"http,omitempty"`
	Chunk        `yaml:"chunk,omitempty"`
	ConvertImage `yaml:"convertImage,omitempty"`
}

// ParseConfigFile is used to parse configuration from yaml file to
// specify configurator
func ParseConfigFile(file string, config *Configurator) error {
	var (
		content []byte
		err     error
	)
	if content, err = ioutil.ReadFile(file); err != nil {
		return err
	}
	return ParseConfig(content, config)
}

// ParseConfig parse config content to specify configurator. If config is nil,
// default global Config will be used
func ParseConfig(configText []byte, config *Configurator) error {
	if config == nil {
		config = DefaultConfig
	}
	return yaml.Unmarshal(configText, config)
}
