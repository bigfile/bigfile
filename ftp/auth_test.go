//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package ftp

import (
	"testing"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/stretchr/testify/assert"
)

func TestAuth_CheckPasswd(t *testing.T) {
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDbConn = trx
	auth := &Auth{}

	_, err = auth.CheckPasswd(tokenPrefix, "")
	assert.Equal(t, ErrTokenNotFound, err)

	secret := models.RandomWithMD5(22)
	token.Secret = &secret
	assert.Nil(t, trx.Model(token).Update("secret", secret).Error)

	_, err = auth.CheckPasswd(tokenPrefix+token.UID, "")
	assert.Equal(t, ErrTokenPassword, err)

	pass, err := auth.CheckPasswd(tokenPrefix+token.UID, secret)
	assert.Nil(t, err)
	assert.True(t, pass)

	_, err = auth.CheckPasswd(token.App.UID, "")
	assert.Equal(t, ErrAppPassword, err)

	pass, err = auth.CheckPasswd(token.App.UID, token.App.Secret)
	assert.Nil(t, err)
	assert.True(t, pass)
}

func TestAuth_CheckPasswd2(t *testing.T) {
	testDbConn = nil
	_, err := (&Auth{}).CheckPasswd(tokenPrefix, "")
	assert.Equal(t, ErrTokenNotFound, err)
}
