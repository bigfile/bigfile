//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"io"
	"os"
	"strings"

	"github.com/bigfile/bigfile/config"
	"github.com/gin-gonic/gin"
)

// Routers will define all routers for service
func Routers() *gin.Engine {
	r := gin.New()
	if config.DefaultConfig.HTTP.AccessLogFile != "" {
		if f, err := os.OpenFile(config.DefaultConfig.HTTP.AccessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			panic(err)
		} else {
			gin.DefaultWriter = io.MultiWriter(os.Stdout, f)
		}
	}
	r.Use(gin.Recovery(), AccessLogMiddleware(), ConfigContextMiddleware(nil), RecordRequestMiddleware())

	requestWithAppGroup := r.Group("", ParseAppMiddleware())
	requestWithAppGroup.POST(
		buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create"),
		SignWithAppMiddleware(&tokenCreateInput{}),
		TokenCreateHandler,
	)

	return r
}

// buildRoute is used to build a correct route
func buildRoute(prefix, route string) string {
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(route, "/")
}
