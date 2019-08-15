//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/bigfile/bigfile/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Routers will define all routers for service
func Routers() *gin.Engine {
	r := gin.New()
	if !isTesting && config.DefaultConfig.HTTP.AccessLogFile != "" {
		setGinLogWriter()
	}

	if config.DefaultConfig.HTTP.CORSEnable {
		r.Use(cors.New(cors.Config{
			AllowAllOrigins:  config.DefaultConfig.CORSAllowAllOrigins,
			AllowOrigins:     config.DefaultConfig.CORSAllowOrigins,
			AllowMethods:     config.DefaultConfig.CORSAllowMethods,
			AllowHeaders:     config.DefaultConfig.CORSAllowHeaders,
			AllowCredentials: config.DefaultConfig.CORSAllowCredentials,
			ExposeHeaders:    config.DefaultConfig.CORSExposeHeaders,
			MaxAge:           time.Duration(config.DefaultConfig.CORSMaxAge * int64(time.Second)),
		}))
	}

	r.Use(gin.Recovery(), AccessLogMiddleware(), ConfigContextMiddleware(nil), RecordRequestMiddleware())

	if !isTesting && config.DefaultConfig.HTTP.LimitRateByIPEnable {
		interval := time.Duration(config.DefaultConfig.HTTP.LimitRateByIPInterval * int64(time.Millisecond))
		maxNumber := config.DefaultConfig.HTTP.LimitRateByIPMaxNum
		r.Use(RateLimitByIPMiddleware(interval, int(maxNumber)))
	}

	requestWithAppGroup := r.Group("", ParseAppMiddleware(), ReplayAttackMiddleware())
	requestWithAppGroup.POST(
		buildRouteWithPrefix("/token/create"),
		SignWithAppMiddleware(&tokenCreateInput{}),
		TokenCreateHandler,
	)
	requestWithAppGroup.POST(
		buildRouteWithPrefix("/token/update"),
		SignWithAppMiddleware(&tokenUpdateInput{}),
		TokenUpdateHandler,
	)

	return r
}

// buildRoute is used to build a correct route
func buildRoute(prefix, route string) string {
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(route, "/")
}

func buildRouteWithPrefix(route string) string {
	return buildRoute(config.DefaultConfig.HTTP.APIPrefix, route)
}

func setGinLogWriter() {
	accessLogFile := config.DefaultConfig.HTTP.AccessLogFile
	dir := filepath.Dir(accessLogFile)
	logger := log.MustNewLogger(&config.DefaultConfig.Log)
	if util.IsFile(dir) {
		logger.Fatalf("invalid access log file path: %s", accessLogFile)
	}
	if !util.IsDir(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Fatal(err)
		}
	}

	if f, err := os.OpenFile(
		config.DefaultConfig.HTTP.AccessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		panic(err)
	} else {
		gin.DefaultWriter = io.MultiWriter(os.Stdout, f)
	}
}
