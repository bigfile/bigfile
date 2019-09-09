//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/gin-gonic/gin"
)

func init() {
	config.DefaultConfig.HTTP.AccessLogFile = ""
	isTesting = true
	gin.SetMode(gin.ReleaseMode)
}

func TestSignatureCalculate(t *testing.T) {
	_ = GetParamsSignature(map[string]interface{}{}, "")
	_ = GetParamsSignBody(map[string]interface{}{}, "")
}
