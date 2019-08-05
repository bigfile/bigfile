//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestToken_TableName(t *testing.T) {
	assert.Equal(t, (&Token{}).TableName(), "tokens")
}

func TestToken_Scope(t *testing.T) {
	token := Token{Path: "/test"}
	assert.Equal(t, token.Scope(), token.Path)
}

func TestNewToken(t *testing.T) {
	var (
		app     *App
		trx     *gorm.DB
		down    func(*testing.T)
		err     error
		token   *Token
		confirm = assert.New(t)
	)
	app, trx, down, err = newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	token, err = NewToken(app, "/test", nil, nil, nil, -1, 0, trx)
	confirm.Equal(err, nil)
	confirm.True(token.ID > 0)
	confirm.Equal(token.Scope(), "/test")
	confirm.Equal(token.Path, "/test")
	confirm.Equal(token.ExpiredAt, (*time.Time)(nil))
	confirm.Equal(token.DeletedAt, (*time.Time)(nil))
	confirm.Equal(token.Secret, (*string)(nil))
	confirm.Equal(token.IP, (*string)(nil))
	confirm.Equal(token.AvailableTimes, -1)
	confirm.Equal(token.ReadOnly, int8(0))
}

func TestNewToken2(t *testing.T) {
	var (
		app         *App
		trx         *gorm.DB
		down        func(*testing.T)
		err         error
		token       *Token
		confirm     = assert.New(t)
		anHourAfter = time.Now().Add(1 * time.Hour)
		ip          = "192.168.1.1"
		secret      = NewSecret()
	)
	app, trx, down, err = newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)

	token, err = NewToken(
		app, "/test", &anHourAfter, &ip, &secret, -1, 0, trx,
	)
	confirm.Equal(err, nil)
	confirm.True(token.ID > 0)
	confirm.Equal(token.Scope(), "/test")
	confirm.Equal(token.Path, "/test")
	confirm.Equal(*token.ExpiredAt, anHourAfter)
	confirm.Equal(token.DeletedAt, (*time.Time)(nil))
	confirm.Equal(*token.Secret, secret)
	confirm.Equal(*token.IP, ip)
}

func TestToken_AllowIPAccess(t *testing.T) {
	var (
		token *Token
		down  func(t *testing.T)
		err   error
		ip    = "192.168.0.1,192.168.0.2"
	)
	confirm := assert.New(t)
	token, _, down, err = newTokenForTest(nil, t, "/test", nil, &ip, nil, -1, int8(0))
	defer down(t)
	confirm.Nil(err)
	confirm.NotNil(token)
	confirm.True(token.AllowIPAccess("192.168.0.2"))
	confirm.False(token.AllowIPAccess("192.168.0.5"))
}

func TestFindTokenByUID(t *testing.T) {
	var (
		token    *Token
		down     func(t *testing.T)
		err      error
		trx      *gorm.DB
		tmpToken *Token
	)
	confirm := assert.New(t)
	token, trx, down, err = newTokenForTest(nil, t, "/test", nil, nil, nil, -1, int8(0))
	defer down(t)
	confirm.Nil(err)
	confirm.NotNil(token)

	tmpToken, err = FindTokenByUID(token.UID, trx)
	confirm.Nil(err)
	confirm.Equal(token.ID, tmpToken.ID)
	confirm.Equal(token.App.ID, tmpToken.App.ID)

	tmpToken, err = FindTokenByUID("a fake token uid", trx)
	confirm.NotNil(err)
	confirm.Contains(err.Error(), "record not found")
}

func TestFindTokenByUIDWithTrashed(t *testing.T) {
	var (
		token    *Token
		down     func(t *testing.T)
		err      error
		trx      *gorm.DB
		tmpToken *Token
	)
	confirm := assert.New(t)
	token, trx, down, err = newTokenForTest(nil, t, "/test", nil, nil, nil, -1, int8(0))
	defer down(t)
	confirm.Nil(err)
	confirm.NotNil(token)

	trx.Delete(token)

	tmpToken, err = FindTokenByUID(token.UID, trx)
	confirm.NotNil(err)
	confirm.Contains(err.Error(), "record not found")

	tmpToken, err = FindTokenByUIDWithTrashed(token.UID, trx)
	confirm.Nil(err)
	confirm.Equal(token.ID, tmpToken.ID)
	confirm.Equal(token.App.ID, tmpToken.App.ID)
}
