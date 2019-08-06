//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenCreate_Validate(t *testing.T) {
	var (
		err         ValidateErrors
		confirm     = assert.New(t)
		tokenCreate *TokenCreate
	)
	tokenCreate = &TokenCreate{AvailableTimes: -2, Path: ""}
	err = tokenCreate.Validate()
	confirm.NotNil(err)
	confirm.True(err.ContainsErrCode(10002))
	confirm.True(err.ContainsErrCode(10003))
	confirm.True(err.ContainsErrCode(10006))
}
