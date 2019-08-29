//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"github.com/bigfile/bigfile/config"
	"github.com/gin-gonic/gin"
)

func init() {
	config.DefaultConfig.HTTP.AccessLogFile = ""
	isTesting = true
	gin.SetMode(gin.ReleaseMode)
}
