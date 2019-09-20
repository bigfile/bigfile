//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
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

	// NewTempDirForTest create a test directory for test
	NewTempDirForTest = newTempDirForTest
)

func setUpTestCaseWithTrx(dbConfig *config.Database, t *testing.T) (*gorm.DB, func(*testing.T)) {
	defer func() { assert.Nil(t, recover()) }()
	db := databases.MustNewConnection(dbConfig)
	trx := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	return trx, func(t *testing.T) {
		defer func() { assert.Nil(t, recover()) }()
		trx.Rollback()
	}
}

func newAppForTest(dbConfig *config.Database, t *testing.T) (*App, *gorm.DB, func(*testing.T), error) {
	defer func() { assert.Nil(t, recover()) }()
	trx, down := setUpTestCaseWithTrx(dbConfig, t)
	note := "test"
	app, err := NewApp("test", &note, trx)
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	token, err = NewToken(app, path, expiredAt, ip, secret, availableTimes, readOnly, trx)
	assert.Nil(t, err)
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
	assert.Nil(t, err)
	token, err = NewToken(app, "/", nil, nil, nil, -1, int8(0), trx)
	assert.Nil(t, err)
	return token, trx, down, err
}

func newTempDirForTest() string {
	return filepath.Join(os.TempDir(), RandomWithMD5(512))
}
