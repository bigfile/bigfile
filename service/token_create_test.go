//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"fmt"
	"testing"

	models "github.com/bigfile/bigfile/databases/mdoels"
)

func TestTokenCreate_Validate(t *testing.T) {
	tokenCreate := &TokenCreate{App: &models.App{}}
	fmt.Println(tokenCreate.Validate())
}
