//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestRecordRequestMiddleware(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	RecordRequestMiddleware()(ctx)
	_, _ = ctx.Writer.Write([]byte("hello"))
	assert.IsType(t, ctx.Writer, &bodyWriter{})
	bw, _ := ctx.Writer.(*bodyWriter)
	assert.Equal(t, "hello", bw.body.String())
}

func TestConfigContextMiddleware(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ConfigContextMiddleware(nil)(ctx)
	dbValue, ok := ctx.Get("db")
	assert.Equal(t, true, ok)
	db, ok := dbValue.(*gorm.DB)
	assert.Equal(t, true, ok)
	assert.NotNil(t, db)
}

func TestParseAppMiddleware(t *testing.T) {
	rc := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rc)
	// appId doesn't set
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	RecordRequestMiddleware()(ctx)
	ConfigContextMiddleware(nil)(ctx)
	ParseAppMiddleware()(ctx)
	bw, _ := ctx.Writer.(*bodyWriter)
	assert.Contains(t, bw.body.String(), "Key: '.AppId' Error:Field validation for 'AppId' failed on the 'required' tag")
}

func TestParseAppMiddleware2(t *testing.T) {
	rc := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rc)
	// input a fake appId
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("appId=fakeAppId"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	RecordRequestMiddleware()(ctx)
	ConfigContextMiddleware(nil)(ctx)
	ParseAppMiddleware()(ctx)
	bw, _ := ctx.Writer.(*bodyWriter)
	assert.Contains(t, bw.body.String(), "cant't parse app from AppId")
}

func TestParseAppMiddleware3(t *testing.T) {
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)

	rc := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rc)
	// input a fake appId
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("appId="+app.UID))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	RecordRequestMiddleware()(ctx)
	ConfigContextMiddleware(trx)(ctx)
	ParseAppMiddleware()(ctx)
	ctxAppValue, ok := ctx.Get("app")
	assert.True(t, ok)
	ctxApp, ok := ctxAppValue.(*models.App)
	assert.True(t, ok)
	assert.Equal(t, app.UID, ctxApp.UID)
}
