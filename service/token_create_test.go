//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"strings"
	"testing"
	"time"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newTokenCreateForTest(t *testing.T) (*TokenCreate, *models.App, func(*testing.T)) {
	var (
		app         *models.App
		err         error
		conn        *gorm.DB
		down        func(*testing.T)
		confirm     = assert.New(t)
		tokenCreate *TokenCreate
		ip          = "192.168.0.1"
		secret      = models.NewSecret()
		expiredAt   = time.Now().Add(100 * time.Second)
	)

	app, conn, down, err = models.NewAppForTest(nil, t)
	confirm.Nil(err)
	tokenCreate = &TokenCreate{
		App:            app,
		AvailableTimes: 1000,
		Path:           "/save/to/test.png",
		IP:             &ip,
		Secret:         &secret,
		ReadOnly:       0,
		ExpiredAt:      &expiredAt,
		BaseService: BaseService{
			DB: conn,
		},
	}
	return tokenCreate, app, down
}

func TestTokenCreate_Validate(t *testing.T) {
	var (
		err         ValidateErrors
		confirm     = assert.New(t)
		tokenCreate *TokenCreate
		ip          string
	)
	ip = strings.Repeat("1", 1501)
	tokenCreate = &TokenCreate{
		AvailableTimes: -2,
		Path:           "",
		IP:             &ip,
		ReadOnly:       2,
	}
	err = tokenCreate.Validate()
	confirm.NotNil(err)
	confirm.True(err.ContainsErrCode(10002))
	confirm.True(err.ContainsErrCode(10003))
	confirm.True(err.ContainsErrCode(10006))
	confirm.True(err.ContainsErrCode(10004))
	confirm.True(err.ContainsErrCode(10007))
}

func TestTokenCreate_Validate2(t *testing.T) {
	tokenCreate, _, down := newTokenCreateForTest(t)
	defer down(t)
	err := tokenCreate.Validate()
	assert.Nil(t, err)
}

func TestTokenCreate_Execute2(t *testing.T) {
	var (
		tokenCreate *TokenCreate
		down        func(*testing.T)
		err         error
		//app         *models.App
		//token       *models.Token
	)
	tokenCreate, _, down = newTokenCreateForTest(t)
	defer down(t)
	err = tokenCreate.Validate()
	assert.Nil(t, err)
	_, err = tokenCreate.Execute(context.TODO())
	assert.Nil(t, err)
}

func TestTokenCreate_Execute(t *testing.T) {
	var (
		tokenCreate *TokenCreate
		down        func(*testing.T)
		err         error
		app         *models.App
		token       *models.Token
		tokenValue  interface{}
		ok          bool
	)
	tokenCreate, app, down = newTokenCreateForTest(t)
	defer down(t)
	err = tokenCreate.Validate()
	assert.Nil(t, err)
	tokenValue, err = tokenCreate.Execute(context.TODO())
	assert.Nil(t, err)
	token, ok = tokenValue.(*models.Token)
	assert.True(t, ok)
	assert.Equal(t, app.ID, token.App.ID)
	assert.True(t, token.ID > 0)
}
