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
	"github.com/stretchr/testify/assert"
)

func TestTokenUpdate_Validate(t *testing.T) {
	var (
		tokenUpdate = &TokenUpdate{
			Token: "",
		}
		err            ValidateErrors
		confirm        = assert.New(t)
		ip             = strings.Repeat("1", 1501)
		path           = strings.Repeat("1", 1001)
		secret         = strings.Repeat("s", 33)
		aSecondAgo     = time.Now().Add(time.Duration(-1 * int64(time.Second)))
		readOnly       = int8(3)
		availableTimes = -2
	)
	tokenUpdate.IP = &ip
	tokenUpdate.Path = &path
	tokenUpdate.ExpiredAt = &aSecondAgo
	tokenUpdate.Secret = &secret
	tokenUpdate.ReadOnly = &readOnly
	tokenUpdate.AvailableTimes = &availableTimes
	err = tokenUpdate.Validate()
	confirm.NotNil(err)
	confirm.True(err.ContainsErrCode(10008))
	confirm.True(err.ContainsErrCode(10009))
	confirm.True(err.ContainsErrCode(10010))
	confirm.True(err.ContainsErrCode(10011))
	confirm.True(err.ContainsErrCode(10012))
	confirm.True(err.ContainsErrCode(10013))
	confirm.True(err.ContainsErrCode(10014))
	confirm.Contains(err.Error(), "path is not a legal unix path")
}

func TestTokenUpdate_Validate2(t *testing.T) {
	var tokenUpdate = &TokenUpdate{
		Token: strings.Repeat("s", 32),
	}
	assert.Nil(t, tokenUpdate.Validate())
}

func TestTokenUpdate_Execute(t *testing.T) {
	expiredAt := time.Now()
	if token, trx, down, err := models.NewTokenForTest(
		nil, t, "/test", &expiredAt, nil, nil, 10, 0); err != nil {
		t.Fatal(err)
	} else {
		defer down(t)
		var (
			tokenUpdate = &TokenUpdate{
				BaseService: BaseService{
					DB: trx,
				},
				Token: token.UID,
			}
			err            ValidateErrors
			confirm        = assert.New(t)
			ip             = strings.Repeat("1", 1500)
			path           = "/new/path"
			secret         = strings.Repeat("s", 32)
			aSecondAgo     = time.Now().Add(time.Hour)
			readOnly       = int8(1)
			availableTimes = 1000
		)
		tokenUpdate.IP = &ip
		tokenUpdate.Path = &path
		tokenUpdate.ExpiredAt = &aSecondAgo
		tokenUpdate.Secret = &secret
		tokenUpdate.ReadOnly = &readOnly
		tokenUpdate.AvailableTimes = &availableTimes
		err = tokenUpdate.Validate()
		confirm.Nil(err)
		if to, err := tokenUpdate.Execute(context.TODO()); err != nil {
			t.Fatal(err)
		} else {
			token, ok := to.(*models.Token)
			confirm.True(ok)
			confirm.Equal(token.AvailableTimes, availableTimes)
			confirm.Equal(token.ReadOnly, readOnly)
			confirm.Equal(token.ExpiredAt.Unix(), aSecondAgo.Unix())
			confirm.Equal(*token.Secret, secret)
			confirm.Equal(token.Path, path)
			confirm.Equal(*token.IP, ip)
		}
	}
}

func TestTokenUpdate_Execute2(t *testing.T) {
	trx, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	var tokenUpdate = &TokenUpdate{
		BaseService: BaseService{
			DB: trx,
		},
		Token: "token does't exist",
	}
	_, err := tokenUpdate.Execute(context.TODO())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "record not found")
}
