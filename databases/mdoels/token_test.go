//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestToken_TableName(t *testing.T) {
	assert.Equal(t, (&Token{}).TableName(), "tokens")
}

func TestToken_Scope(t *testing.T) {
	token := Token{Path: "/test"}
	assert.Equal(t, token.Scope(), token.Path)
}
