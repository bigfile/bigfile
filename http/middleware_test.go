//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
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
	// appUid doesn't set
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
	assert.Contains(t, bw.body.String(), "Error:Field validation for 'AppUID' failed on the 'required' tag")
}

func TestParseAppMiddleware2(t *testing.T) {
	rc := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rc)
	// input a fake appUid
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("appUid=fakeAppUid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx.Request = req
	db, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = db
	ConfigContextMiddleware(nil)(ctx)
	RecordRequestMiddleware()(ctx)
	ParseAppMiddleware()(ctx)
	bw, _ := ctx.Writer.(*bodyWriter)
	assert.Contains(t, bw.body.String(), "cant't parse app from appUid")
}

func TestParseAppMiddleware3(t *testing.T) {
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)

	rc := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rc)
	// input a fake appUid
	req, _ := http.NewRequest("POST", "http://bigfile.io",
		strings.NewReader("appUid="+app.UID))
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
		strings.NewReader("appUid=fakeAppUid"))
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

func TestRateLimitByIPMiddleware(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader(""))
	ctx.Request.Header.Set("X-Forwarded-For", "192.168.0.1")
	ctx.Set("requestId", int64(10000001111))
	bw := &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Writer = bw

	RateLimitByIPMiddleware(time.Second, 1)(ctx)
	assert.Equal(t, 0, bw.body.Len())
	RateLimitByIPMiddleware(time.Second, 1)(ctx)
	assert.Equal(t, 429, ctx.Writer.Status())
	assert.Contains(t, bw.body.String(), "too many requests")
}

func TestReplayAttackMiddleware(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	bw := &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Writer = bw
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader("nonce=1111"))
	ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	app, db, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	reqRecord := models.MustNewRequestWithProtocol("http", db)
	ctx.Set("db", db)
	ctx.Set("app", app)
	ctx.Set("reqRecord", reqRecord)
	ReplayAttackMiddleware()(ctx)
	assert.Equal(t, 400, ctx.Writer.Status())
	assert.Contains(t, bw.body.String(), "nonce is optional, but the min length of nonce is 32, the max length is 48")
}

func TestReplayAttackMiddleware2(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	bw := &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Writer = bw
	nonce := models.RandomWithMd5(128)
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader("nonce="+nonce))
	ctx.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	app, db, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	reqRecord := models.MustNewRequestWithProtocol("http", db)
	reqRecord.Nonce = &nonce
	reqRecord.AppID = &app.ID
	assert.Nil(t, reqRecord.Save(db))

	ctx.Set("db", db)
	ctx.Set("app", app)
	ctx.Set("reqRecord", reqRecord)

	ReplayAttackMiddleware()(ctx)
	assert.Equal(t, 400, ctx.Writer.Status())
	assert.Contains(t, bw.body.String(), "this request is being replayed")
}

func TestBodyWriter_Write(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	bw := &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Writer = bw
	n, err := bw.Write([]byte("hello"))
	assert.Nil(t, err)
	assert.True(t, n == 5)
}
