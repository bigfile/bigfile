//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newTokenDeleteForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
	var (
		ctx   *gin.Context
		trx   *gorm.DB
		err   error
		token *models.Token
		down  func(*testing.T)
	)
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Writer = &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Request, _ = http.NewRequest("PATCH", "http://bigfile.io", strings.NewReader(""))
	ctx.Request.Header.Set("X-Forwarded-For", "192.168.0.1")
	ctx.Set("app", &token.App)
	ctx.Set("db", trx)
	ctx.Set("token", token)
	reqRecord := models.MustNewRequestWithProtocol("http", trx)
	ctx.Set("reqRecord", reqRecord)
	ctx.Set("requestId", int64(reqRecord.ID))

	ctx.Set("inputParam", &tokenDeleteInput{
		Token: token.UID,
	})

	return ctx, func(t *testing.T) {
		down(t)
	}
}

func TestTokenDeleteHandler(t *testing.T) {
	ctx, down := newTokenDeleteForTest(t)
	defer down(t)
	input := ctx.MustGet("inputParam").(*tokenDeleteInput)
	input.Token = ""
	writer := ctx.Writer.(*bodyWriter)
	TokenDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "record not found", response.Errors["token"][0])
}

func TestTokenDeleteHandler2(t *testing.T) {
	ctx, down := newTokenDeleteForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)
	TokenDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, responseData["deletedAt"])
}

func TestTokenDeleteHandler3(t *testing.T) {
	var (
		w     = httptest.NewRecorder()
		err   error
		trx   *gorm.DB
		api   = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/delete")
		token *models.Token
		down  func(*testing.T)
	)
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx

	params := map[string]interface{}{
		"appUid": token.App.UID,
		"token":  token.UID,
		"nonce":  models.RandomWithMD5(222),
	}

	apiWithQs := fmt.Sprintf("%s?%s", api, getParamsSignBody(params, token.App.Secret))
	req, _ := http.NewRequest("DELETE", apiWithQs, nil)
	Routers().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, err)
	response, err := parseResponse(w.Body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, responseData["deletedAt"])
}

func BenchmarkTokenDeleteHandler(b *testing.B) {
	b.StopTimer()

	var (
		api = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/delete")
		trx *gorm.DB
		err error
	)

	trx = databases.MustNewConnection(nil).BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	testDBConn = trx
	defer func() {
		trx.Rollback()
	}()

	note := "test"
	app, err := models.NewApp("test", &note, trx)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		var token *models.Token
		if token, err = models.NewToken(app, "/benchmark", nil, nil, nil, 100, 0, trx); err != nil {
			b.Fatal(err)
		}
		var (
			w      = httptest.NewRecorder()
			params = map[string]interface{}{
				"appUid": token.App.UID,
				"token":  token.UID,
				"nonce":  models.RandomWithMD5(222),
			}
		)
		apiWithQs := fmt.Sprintf("%s?%s", api, getParamsSignBody(params, app.Secret))
		req, _ := http.NewRequest("DELETE", apiWithQs, nil)
		Routers().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatal("response code should be 200")
		}
	}
}
