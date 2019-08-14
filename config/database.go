//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
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
