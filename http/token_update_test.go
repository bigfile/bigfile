//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestTokenUpdateHandler is used to run TokenUpdateHandler directly
func TestTokenUpdateHandler(t *testing.T) {
	token, trx, down, err := models.NewTokenForTest(
		nil, t, "/test", nil, nil, nil, 1000, 0)
	assert.Nil(t, err)
	defer down(t)

	var (
		path           = "/to"
		ip             = "192.168.0.1"
		expiredAt      = time.Now().Add(1 * time.Hour)
		secret         = models.NewSecret()
		availableTimes = 100
		readOnly       = true
		input          = &tokenUpdateInput{
			Token:          token.UID,
			Path:           &path,
			IP:             &ip,
			ExpiredAt:      &expiredAt,
			Secret:         &secret,
			AvailableTimes: &availableTimes,
			ReadOnly:       &readOnly,
		}
	)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	bw := &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Writer = bw
	ctx.Set("db", trx)
	ctx.Set("requestId", rand.Int63n(100000000))
	ctx.Set("inputParam", input)
	TokenUpdateHandler(ctx)
	response, err := parseResponse(bw.body.String())
	assert.Nil(t, err)
	assertTokenRespStructure(response.Data)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, responseData["path"].(string), path)
	assert.Equal(t, responseData["ip"].(string), ip)
	assert.Equal(t, responseData["secret"].(string), secret)
	assert.Equal(t, int64(responseData["expiredAt"].(float64)), expiredAt.Unix())
	assert.Equal(t, int(responseData["availableTimes"].(float64)), availableTimes)
	assert.Equal(t, int(responseData["readOnly"].(float64)), 1)
}

// TestTokenUpdateHandler2 Actually, is used to not do any change to token
func TestTokenUpdateHandler2(t *testing.T) {
	token, trx, down, err := models.NewTokenForTest(
		nil, t, "/test", nil, nil, nil, 1000, 0)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx

	var (
		app  = token.App
		w    = httptest.NewRecorder()
		api  = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/update")
		body = fmt.Sprintf("appUid=%s&nonce=%s&token=%s",
			app.UID, models.RandomWithMd5(128), token.UID)
		sign = SignStrWithSecret(body, app.Secret)
	)
	body = fmt.Sprintf("%s&sign=%s", body, sign)
	req, _ := http.NewRequest("POST", api, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	Routers().ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	response, err := parseResponse(w.Body.String())
	assert.Nil(t, err)
	assertTokenRespStructure(response.Data)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, responseData["path"].(string), "/test")
	assert.Nil(t, responseData["ip"])
	assert.Nil(t, responseData["secret"])
	assert.Nil(t, responseData["expiredAt"])
	assert.Equal(t, int(responseData["availableTimes"].(float64)), 1000)
	assert.Equal(t, int(responseData["readOnly"].(float64)), 0)
}

func TestTokenUpdateHandler3(t *testing.T) {
	token, trx, down, err := models.NewTokenForTest(
		nil, t, "/test", nil, nil, nil, 1000, 0)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx

	var (
		app       = token.App
		w         = httptest.NewRecorder()
		api       = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/update")
		expiredAt = time.Now().Add(time.Hour).Unix()
		secret    = models.NewSecret()
		body      = getParamsSignBody(map[string]interface{}{
			"appUid":         app.UID,
			"nonce":          models.RandomWithMd5(32),
			"token":          token.UID,
			"expiredAt":      expiredAt,
			"ip":             "192.168.0.1",
			"secret":         secret,
			"availableTimes": 100,
			"readOnly":       1,
			"path":           "hello",
		}, app.Secret)
	)
	req, _ := http.NewRequest("POST", api, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	Routers().ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	response, err := parseResponse(w.Body.String())
	assert.Nil(t, err)
	assertTokenRespStructure(response.Data)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, responseData["path"].(string), "/hello")
	assert.Equal(t, responseData["ip"].(string), "192.168.0.1")
	assert.Equal(t, responseData["secret"].(string), secret)
	assert.Equal(t, int64(responseData["expiredAt"].(float64)), expiredAt)
	assert.Equal(t, int(responseData["availableTimes"].(float64)), 100)
	assert.Equal(t, int(responseData["readOnly"].(float64)), 1)
}

func BenchmarkTokenUpdateHandler(b *testing.B) {
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

	token, err := models.NewToken(app, "/", nil, nil, nil, 100, 0, trx)
	if err != nil {
		b.Fatal(err)
	}

	var (
		api       = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/update")
		expiredAt = time.Now().Add(time.Hour).Unix()
		secret    = models.NewSecret()
		routers   = Routers()
	)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		body := getParamsSignBody(map[string]interface{}{
			"appUid":         app.UID,
			"nonce":          models.RandomWithMd5(32),
			"token":          token.UID,
			"expiredAt":      expiredAt,
			"ip":             "192.168.0.1",
			"secret":         secret,
			"availableTimes": 100,
			"readOnly":       1,
			"path":           "hello",
		}, app.Secret)
		req, _ := http.NewRequest("POST", api, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		routers.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatal("the code of request should be 200")
		}
	}
}
