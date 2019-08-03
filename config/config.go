//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package config is responsible parsing the configuration file,
// and then, it is used by other parts of application.
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
)

// Database represent database config
type Database struct {

	// driver can be sqlite3, mysql or postgres
	Driver string `yaml:"driver,omitempty"`

	// mysql and postgres database config
	// database server host
	Host string `yaml:"host,omitempty"`

	// database user
	User string `yaml:"user,omitempty"`

	// database password
	Password string `yaml:"password,omitempty"`

	// database name
	DBName string `yaml:"dbName,omitempty"`

	// database server port
	Port uint32 `yaml:"port,omitempty"`

	// sqlite3 db file
	DBFile string `yaml:"dbFile,omitempty"`
}

// ConsoleLog is used to config output handle for stdout
type ConsoleLog struct {
	// Enable represent enable or disable log
	Enable bool `yaml:"enable,omitempty"`

	// Level represent stdout log level
	Level string `yaml:"level,omitempty"`

	// Format represent log format for stdout
	Format string `yaml:"format,omitempty"`

	// output represent destination path of log output, default: stdout
	output *os.File `yaml:"-"`
}

// SetOutput is used to set output path of log
func (c *ConsoleLog) SetOutput(output *os.File) {
	c.output = output
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

// DSN generate dsn based on db driver
func (d Database) DSN() (string, error) {
	switch d.Driver {
	case "sqlite3":
		return fmt.Sprintf("file:%s?mode=rw&cache=shared", d.DBFile), nil
	case "mysql":
		format := "%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local"
		return fmt.Sprintf(format, d.User, d.Password, d.Host, d.Port, d.DBName), nil
	case "postgres":
		format := "host=%s port=%d user=%s dbname=%s password=%s"
		return fmt.Sprintf(format, d.Host, d.Port, d.User, d.DBName, d.Password), nil
	default:
		return "", errors.New("unsupported database driver")
	}
}

// Configurator combine every child config
type Configurator struct {
	Database `yaml:"database,omitempty"`
	Log      `yaml:"log,omitempty"`
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
				Path:            "bigfile.log",
				MaxBytesPerFile: 52428800,
			},
		},
	}
)

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
