//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"testing"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/jinzhu/gorm"
)

// Down represent rollback function type
type Down = func(*testing.T)

var (
	// NewAppForTest export newAppForTest for other package
	NewAppForTest = newAppForTest

	// SetUpTestCaseWithTrx is a helper method for helping to finish test
	SetUpTestCaseWithTrx = setUpTestCaseWithTrx

	// NewTokenForTest export newTokenForTest for other package
	NewTokenForTest = newTokenForTest

	// NewArbitrarilyTokenForTest export newArbitrarilyTokenForTest
	NewArbitrarilyTokenForTest = newArbitrarilyTokenForTest
)

func setUpTestCaseWithTrx(dbConfig *config.Database, t *testing.T) (*gorm.DB, func(*testing.T)) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	db := databases.MustNewConnection(dbConfig)
	trx := db.Begin()
	return trx, func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Fatal(err)
			}
		}()
		trx.Rollback()
	}
}

func newAppForTest(dbConfig *config.Database, t *testing.T) (*App, *gorm.DB, func(*testing.T), error) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	trx, down := setUpTestCaseWithTrx(dbConfig, t)
	note := "test"
	app, err := NewApp("test", &note, trx)
	if err != nil {
		t.Fatal(err)
	}
	return app, trx, down, err
}

func newTokenForTest(
	cfg *config.Database,
	t *testing.T,
	path string,
	expiredAt *time.Time,
	ip, secret *string,
	availableTimes int,
	readOnly int8,
) (*Token, *gorm.DB, func(*testing.T), error) {
	var (
		app   *App
		trx   *gorm.DB
		down  func(*testing.T)
		err   error
		token *Token
	)
	app, trx, down, err = newAppForTest(cfg, t)
	if err != nil {
		t.Fatal(err)
	}
	token, err = NewToken(app, path, expiredAt, ip, secret, availableTimes, readOnly, trx)
	if err != nil {
		t.Fatal(err)
	}
	return token, trx, down, err
}

func newArbitrarilyTokenForTest(cfg *config.Database, t *testing.T) (*Token, *gorm.DB, func(*testing.T), error) {
	var (
		app   *App
		trx   *gorm.DB
		down  func(*testing.T)
		err   error
		token *Token
	)
	app, trx, down, err = newAppForTest(cfg, t)
	if err != nil {
		t.Fatal(err)
	}
	token, err = NewToken(app, "/", nil, nil, nil, -1, int8(0), trx)
	if err != nil {
		t.Fatal(err)
	}
	return token, trx, down, err
}
