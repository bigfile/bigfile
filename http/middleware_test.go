//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestRecordRequestMiddleware(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader(""))
	db, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = db
	ConfigContextMiddleware(nil)(ctx)
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
	db, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = db
	ConfigContextMiddleware(nil)(ctx)
	RecordRequestMiddleware()(ctx)
	ConfigContextMiddleware(nil)(ctx)
	ParseAppMiddleware()(ctx)
	bw, _ := ctx.Writer.(*bodyWriter)
	assert.Contains(t, bw.body.String(), "Key: '.AppID' Error:Field validation for 'AppID' failed on the 'required' tag")
}

func TestParseAppMiddleware2(t *testing.T) {
	rc := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rc)
	// input a fake appId
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("appId=fakeAppId"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	db, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = db
	ConfigContextMiddleware(nil)(ctx)
	RecordRequestMiddleware()(ctx)
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
	db, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = db
	ConfigContextMiddleware(trx)(ctx)
	RecordRequestMiddleware()(ctx)
	ParseAppMiddleware()(ctx)
	ctxAppValue, ok := ctx.Get("app")
	assert.True(t, ok)
	ctxApp, ok := ctxAppValue.(*models.App)
	assert.True(t, ok)
	assert.Equal(t, app.UID, ctxApp.UID)
}

func TestValidateRequestSignature(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// no input sign param
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("appId=fakeAppId"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	assert.Nil(t, ctx.Request.ParseForm())
	assert.False(t, ValidateRequestSignature(ctx, "secret"))

	m := md5.New()
	secret := "this is a test secret"
	body := "a=1&b=2&c=3&d=4"
	_, _ = m.Write([]byte(body))
	_, _ = m.Write([]byte(secret))
	correctSign := hex.EncodeToString(m.Sum(nil))

	// sign error
	bodyWithWrongSign := body + "&sign=" + "wrong sign"
	req, _ = http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader(bodyWithWrongSign))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	assert.Nil(t, ctx.Request.ParseForm())
	assert.False(t, ValidateRequestSignature(ctx, secret))

	// sign correctly
	bodyWithWrongSign = body + "&sign=" + correctSign
	req, _ = http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader(bodyWithWrongSign))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	assert.Nil(t, ctx.Request.ParseForm())
	assert.True(t, ValidateRequestSignature(ctx, secret))
}

func TestSignWithAppMiddleware(t *testing.T) {

	type TestInput struct {
		A string     `form:"a" json:"a" binding:"required"`
		B int64      `form:"b" json:"b" binding:"required"`
		C bool       `form:"c" json:"c" binding:"required"`
		T *time.Time `form:"t" json:"t"`
	}

	var it = &TestInput{}

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader(""))
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx
	ConfigContextMiddleware(nil)(ctx)
	RecordRequestMiddleware()(ctx)
	assert.IsType(t, ctx.Writer, &bodyWriter{})
	bw, _ := ctx.Writer.(*bodyWriter)

	// parse param error, because of c type error
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("a=a&b=1&c=c"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	SignWithAppMiddleware(it)(ctx)
	assert.Contains(t, bw.body.String(), "strconv.ParseBool: parsing")
	bw.body.Reset()

	// sign error
	ctx.Set("app", app)
	req, _ = http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("a=a&b=1&c=1&sign=error"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req

	SignWithAppMiddleware(it)(ctx)
	assert.Contains(t, bw.body.String(), "request param sign error")
	bw.body.Reset()

	// sign correctly
	reqBody := "a=a&b=1&c=1"
	reqBodySign := SignStrWithSecret(reqBody, app.Secret)
	reqBody += "&sign=" + reqBodySign
	req, _ = http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	SignWithAppMiddleware(it)(ctx)
	assert.Equal(t, 0, bw.body.Len())
}

func TestSignStrWithSecret(t *testing.T) {
	assert.Equal(t, SignStrWithSecret("a=1&b=2&c=3", "secret"), "44d18b899e05f82c1cc4ce22bf5df09b")
}
