//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package ftp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverFactory_NewDriver(t *testing.T) {
	_, err := (&Factory{}).NewDriver()
	assert.Nil(t, err)
}
