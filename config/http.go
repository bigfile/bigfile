//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package config

// HTTP define config format for http service
type HTTP struct {

	// APIPrefix represent api prefix for each route.
	// example: /api/v1
	APIPrefix string `yaml:"apiPrefix,omitempty"`

	// AccessLogFile represent http access log file. if it's empty,
	// log will not be written to accessLogFile
	AccessLogFile string `yaml:"accessLogFile,omitempty"`

	// LimitRateByIPEnable represents whether enable limit rate middleware
	// default: false
	LimitRateByIPEnable bool `yaml:"limitRateByIPEnable,omitempty"`

	// LimitRateInterval represent limit interval, unit: ms, default: 1000ms, that is 1s
	LimitRateByIPInterval int64 `yaml:"limitRateByIPInterval,omitempty"`

	// LimitRateByIPMaxNum represent max request limit per LimitRateByIPInterval
	// default: 100
	LimitRateByIPMaxNum uint `yaml:"limitRateByIPMaxNum,omitempty"`
}
