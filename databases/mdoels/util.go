//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/bigfile/bigfile/databases"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/log"
)

// NewSecret is used generate secret for app and token
func NewSecret() string {
	var (
		b           = make([]byte, 128)
		hash        = md5.New()
		letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)
	if _, err := rand.Reader.Read(b); err != nil {
		log.MustNewLogger(&config.DefaultConfig.Log).Warning(err)
		for i := range b {
			b[i] = letterBytes[mrand.Intn(len(letterBytes))]
		}
	}
	_, _ = hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

func setUpTestCaseWithTrx(dbConfig *config.Database, t *testing.T) (*gorm.DB, func(*testing.T)) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	db := databases.MustNewConnection(dbConfig, true)
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
