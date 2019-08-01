//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package databases provides capacity to interact with database
package databases

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/jinzhu/gorm"

	// import mysql database driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
	// import postgres database driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// import sqlite database driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	connection *gorm.DB
	pool       []*gorm.DB
)

// NewConnection will initialize a connection to database. But, if connection
// has already existed, it will be used, except `useCache` is false. If something
// goes wrong, an error will be used to represent it.
func NewConnection(dbConfig *config.Database, useCache bool) (*gorm.DB, error) {
	if dbConfig == nil {
		dbConfig = &config.DefaultConfig.Database
	}
	if useCache && connection != nil {
		return connection, nil
	}

	dsn, err := dbConfig.DSN()
	if err != nil {
		return nil, err
	}

	conn, err := gorm.Open(dbConfig.Driver, dsn)
	if err != nil {
		return nil, err
	}

	if connection == nil {
		connection = conn
		// ensure default db connection is alive always, avoid being automatically disconnected
		go func() {
			connection.Raw("select 1")
			time.Sleep(10 * time.Minute)
		}()
	}
	pool = append(pool, conn)

	// Register the exit signal handler, ensure to close all alive connection
	if len(pool) == 1 {
		quitSignal := make(chan os.Signal, 1)
		signal.Notify(quitSignal, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGINT)
		go func() {
			<-quitSignal
			for _, conn := range pool {
				// ignore the close error, maybe the connection has been already closed
				fmt.Println("disconnect database")
				_ = conn.Close()
			}
		}()
		signal.Stop(quitSignal)
	}
	return conn, nil
}
