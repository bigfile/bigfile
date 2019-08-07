//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package config

import (
	"github.com/op/go-logging"
)

// ConsoleLog is used to config output handle for stdout
type ConsoleLog struct {
	// Enable represent enable or disable log
	Enable bool `yaml:"enable,omitempty"`

	// Level represent stdout log level
	Level string `yaml:"level,omitempty"`

	// Format represent log format for stdout
	Format string `yaml:"format,omitempty"`
}

// FileLog is used to config output handler for file
type FileLog struct {
	// Enable represent enable or disable log
	Enable bool `yaml:"enable,omitempty"`

	// Level represent file log level
	Level string `yaml:"level,omitempty"`

	// Format represent log format for file
	Format string `yaml:"format,omitempty"`

	// Path is used to config log output destination
	Path string `yaml:"path,omitempty"`

	// MaxBytesPerFile is used to config max capacity of single log file, unit: bytes
	MaxBytesPerFile uint64 `yaml:"maxBytesPerFile,omitempty"`
}

// Log includes configuration item for log component
type Log struct {
	Console ConsoleLog `yaml:"console,omitempty"`
	File    FileLog    `yaml:"file,omitempty"`
}

var (
	// LevelToName map logging.Level to string represent
	LevelToName = map[logging.Level]string{
		logging.DEBUG:    "DEBUG",
		logging.INFO:     "INFO",
		logging.NOTICE:   "NOTICE",
		logging.WARNING:  "WARNING",
		logging.ERROR:    "ERROR",
		logging.CRITICAL: "CRITICAL",
	}
	// NameToLevel map string represent of log level to logging.Level
	NameToLevel = map[string]logging.Level{
		"DEBUG":    logging.DEBUG,
		"INFO":     logging.INFO,
		"NOTICE":   logging.NOTICE,
		"WARNING":  logging.WARNING,
		"ERROR":    logging.ERROR,
		"CRITICAL": logging.CRITICAL,
	}
)
