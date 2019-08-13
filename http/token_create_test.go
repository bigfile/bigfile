//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bigfile/bigfile/databases"

	"github.com/bigfile/bigfile/config"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/gin-gonic/gin"
	janitor "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestTokenCreateHandler(t *testing.T) {
	var (
		readOnly       bool
		availableTimes = 1
		path           = "/"
		input          = &tokenCreateInput{
			Nonce:          models.RandomWithMd5(128),
			Path:           &path,
			AvailableTimes: &availableTimes,
			ReadOnly:       &readOnly,
		}
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader(""))
	db, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = db
	ConfigContextMiddleware(nil)(ctx)
	RecordRequestMiddleware()(ctx)
	assert.IsType(t, ctx.Writer, &bodyWriter{})
	bw, _ := ctx.Writer.(*bodyWriter)
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	ctx.Set("app", app)
	ctx.Set("inputParam", input)
	ctx.Set("db", trx)
	TokenCreateHandler(ctx)
	assert.Equal(t, 200, ctx.Writer.Status())

	var response = &Response{}
	json := janitor.ConfigCompatibleWithStandardLibrary
	assert.Nil(t, json.Unmarshal(bw.body.Bytes(), response))
	assert.True(t, response.Success)
	respData := response.Data.(map[string]interface{})
	respAvailableTimes := respData["availableTimes"].(float64)
	assert.Equal(t, availableTimes, int(respAvailableTimes))
}

func TestTokenCreateHandler2(t *testing.T) {
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx
	router := Routers()

	w := httptest.NewRecorder()
	api := buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create")
	body := fmt.Sprintf("appUid=%s&nonce=%s", app.UID, models.RandomWithMd5(128))
	sign := SignStrWithSecret(body, app.Secret)
	body = fmt.Sprintf("%s&sign=%s", body, sign)
	req, _ := http.NewRequest("POST", api, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var response = &Response{}
	json := janitor.ConfigCompatibleWithStandardLibrary
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), response))
	assert.True(t, response.Success)

	respData := response.Data.(map[string]interface{})
	respAvailableTimes := respData["availableTimes"].(float64)
	assert.Equal(t, -1, int(respAvailableTimes))
	respTokenValue := respData["token"].(string)
	assert.Equal(t, 24, len(respTokenValue))
	assert.Nil(t, respData["ip"])
	respReadOnly := respData["readOnly"].(float64)
	assert.Equal(t, 0, int(respReadOnly))
	respReadPath := respData["path"].(string)
	assert.Equal(t, "/", respReadPath)
	assert.Nil(t, respData["expiredAt"])
}

func TestTokenCreateHandler3(t *testing.T) {
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx
	router := Routers()

	w := httptest.NewRecorder()
	api := buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create")
	expiredAt := time.Now().Add(10 * time.Hour)
	expiredAtUnix := expiredAt.Unix()
	secret := SignStrWithSecret("", "")
	body := fmt.Sprintf(
		"appUid=%s&availableTimes=1000&expiredAt=%d&ip=192.168.0.1&nonce=%s&path=/test&readOnly=1&secret=%s",
		app.UID, expiredAtUnix, models.RandomWithMd5(128), secret,
	)
	sign := SignStrWithSecret(body, app.Secret)
	body = fmt.Sprintf("%s&sign=%s", body, sign)
	req, _ := http.NewRequest("POST", api, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var response = &Response{}
	json := janitor.ConfigCompatibleWithStandardLibrary
	assert.Nil(t, json.Unmarshal(w.Body.Bytes(), response))
	assert.True(t, response.Success)

	respData := response.Data.(map[string]interface{})
	respAvailableTimes := respData["availableTimes"].(float64)
	assert.Equal(t, 1000, int(respAvailableTimes))
	respTokenValue := respData["token"].(string)
	assert.Equal(t, 24, len(respTokenValue))
	respIP := respData["ip"].(string)
	assert.Equal(t, "192.168.0.1", respIP)
	respReadOnly := respData["readOnly"].(float64)
	assert.Equal(t, 1, int(respReadOnly))
	respReadPath := respData["path"].(string)
	assert.Equal(t, "/test", respReadPath)
	respExpiredAt := respData["expiredAt"].(string)
	respExpiredAtTime, _ := time.Parse(time.RFC3339, respExpiredAt)
	assert.Equal(t, respExpiredAtTime.Unix(), expiredAtUnix)
}

func BenchmarkTokenCreateHandler(b *testing.B) {
	b.StopTimer()
	defer func() {
		if err := recover(); err != nil {
			b.Fatal(err)
		}
	}()

	trx := databases.MustNewConnection(nil).Begin()
	testDBConn = trx
	defer func() { trx.Rollback() }()

	note := "test"
	app, err := models.NewApp("test", &note, trx)
	if err != nil {
		b.Fatal(err)
	}

	router := Routers()
	w := httptest.NewRecorder()
	api := buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create")
	expiredAt := time.Now().Add(10 * time.Hour)
	expiredAtUnix := expiredAt.Unix()
	secret := SignStrWithSecret("", "")
	body := fmt.Sprintf(
		"appUid=%s&availableTimes=1000&expiredAt=%d&ip=192.168.0.1&nonce=%s&path=/test&readOnly=1&secret=%s",
		app.UID, expiredAtUnix, models.RandomWithMd5(128), secret,
	)
	sign := SignStrWithSecret(body, app.Secret)
	body = fmt.Sprintf("%s&sign=%s", body, sign)
	req, _ := http.NewRequest("POST", api, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			b.Fatalf("response code should be 200")
		}
	}
}

func TestRateLimitByIPMiddleware2(t *testing.T) {
	paramStr := fmt.Sprintf("appUid=5d4ee21527075f8740000001&nonce=%s&path=test", models.RandomWithMd5(128))
	fmt.Println(paramStr)
	fmt.Println(SignStrWithSecret(paramStr, "a0ee01d762d9d167cbfa5656614e62d7"))
}
