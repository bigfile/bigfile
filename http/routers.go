//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/gin-gonic/gin"
)

var cfg = config.DefaultConfig

// buildRoute is used to build a correct route
func buildRoute(prefix, route string) string {
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(route, "/")
}

// Routers will define all routers for service
func Routers() *gin.Engine {

	r := gin.New()
	r.POST(buildRoute(cfg.HTTP.APIPrefix, "/token/create"), TokenCreateHandler)

	if cfg.HTTP.AccessLogFile != "" {
		if f, err := os.OpenFile(cfg.HTTP.AccessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			panic(err)
		} else {
			gin.DefaultWriter = io.MultiWriter(os.Stdout, f)
		}
	}

	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	r.Use(gin.Recovery())

	return r
}
