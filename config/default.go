//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package config

import (
	"time"

	"github.com/op/go-logging"
)

var (
	// DefaultConfig define a default configurator
	DefaultConfig = &Configurator{
		Database{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     3306,
			User:     "root",
			Password: "root",
			DBName:   "bigfile",
		},
		Log{
			Console: ConsoleLog{
				Level:  LevelToName[logging.DEBUG],
				Enable: true,
				Format: `%{color:bold}[%{time:2006/01/02 15:04:05.000}] %{pid} %{level:.5s} %{color:reset} %{message}`,
			},
			File: FileLog{
				Enable:          true,
				Level:           LevelToName[logging.WARNING],
				Format:          "[%{time:2006/01/02 15:04:05.000}] %{pid} %{longfile} %{longfunc} %{callpath} ▶ %{level:.4s} %{message}",
				Path:            "storage/logs/bigfile.log",
				MaxBytesPerFile: 52428800,
			},
		},
		HTTP{
			APIPrefix:             "/api/bigfile",
			AccessLogFile:         "storage/logs/bigfile.http.access.log",
			LimitRateByIPEnable:   false,
			LimitRateByIPInterval: 1000,
			LimitRateByIPMaxNum:   100,
			CORSEnable:            false,
			CORSAllowAllOrigins:   false,
			CORSAllowCredentials:  false,
			CORSAllowHeaders:      []string{"Origin"},
			CORSAllowOrigins:      []string{"*"},
			CORSAllowMethods:      []string{"PUT", "DELETE"},
			CORSMaxAge:            3600 * int64(time.Second),
		},
	}
)
