//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import "github.com/gin-gonic/gin"

func init() {
	isTesting = true
	gin.SetMode(gin.ReleaseMode)
}
