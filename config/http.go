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
}
