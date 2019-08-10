//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var config = `database:
  driver: mysql
  host: localhost
  user: root
  password: root
  port: 3306
  dbName: bigfile
log:
  console:
    enable: true
    level: info
    format: '%{color:bold}[%{time:2006/01/02 15:04:05.000}] %{pid} %{level:.5s} %{color:reset} %{message}'
  file:
    enable: true
    path: 'bigfile.log'
    level: warn
    format: '[%{time:2006/01/02 15:04:05.000}] %{pid} %{longfile} %{longfunc} %{callpath} ▶ %{level:.4s} %{message}'
    maxBytesPerFile: 52428800
http:
  apiPrefix: /api/bigfile`

func assertConfigurator(t *testing.T, configurator *Configurator) {
	confirm := assert.New(t)
	confirm.Equal("mysql", configurator.Driver)
	confirm.Equal("localhost", configurator.Host)
	confirm.Equal("root", configurator.User)
	confirm.Equal("root", configurator.Password)
	confirm.Equal(uint32(3306), configurator.Port)
	confirm.Equal("bigfile", configurator.DBName)
	confirm.Equal(true, configurator.Log.Console.Enable)
	confirm.Equal("info", configurator.Log.Console.Level)
	confirm.Equal(
		"%{color:bold}[%{time:2006/01/02 15:04:05.000}] %{pid} %{level:.5s} %{color:reset} %{message}",
		configurator.Log.Console.Format,
	)
	confirm.Equal("bigfile.log", configurator.File.Path)
	confirm.Equal("warn", configurator.File.Level)
	confirm.Equal(true, configurator.Log.File.Enable)
	confirm.Equal(uint64(52428800), configurator.File.MaxBytesPerFile)
	confirm.Equal(
		"[%{time:2006/01/02 15:04:05.000}] %{pid} %{longfile} %{longfunc} %{callpath} ▶ %{level:.4s} %{message}",
		configurator.File.Format,
	)
	confirm.Equal("/api/bigfile", configurator.HTTP.APIPrefix)
}

func TestParseConfigFile(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile(os.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpConfigFile.Name())
	if _, err := tmpConfigFile.Write([]byte(config)); err != nil {
		t.Fatal(err)
	}
	configurator := &Configurator{}
	if err := ParseConfigFile(tmpConfigFile.Name(), configurator); err != nil {
		t.Fatal(err)
	}
	assertConfigurator(t, configurator)
}

func TestParseConfig(t *testing.T) {
	configurator := &Configurator{}
	if err := ParseConfig([]byte(config), configurator); err != nil {
		t.Fatal(err)
	}
	assertConfigurator(t, configurator)
}

func TestDatabase_DSN(t *testing.T) {
	configurator := &Configurator{}
	if err := ParseConfig([]byte(config), configurator); err != nil {
		t.Fatal(err)
	}
	if _, err := configurator.DSN(); err != nil {
		t.Fatal(err)
	}
	configurator.Driver = "unknown driver"
	_, err := configurator.DSN()
	assert.NotEqual(t, err, nil)
	assert.Contains(t, err.Error(), "unsupported database driver")
}
