//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"net/http/httptest"
	"testing"
	"time"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestTokenCreateHandler(t *testing.T) {
	var (
		readOnly       bool
		availableTimes = 1
		path           = "/"
		input          = &tokenCreateInput{
			RequestTime:    time.Now(),
			Path:           &path,
			AvailableTimes: &availableTimes,
			ReadOnly:       &readOnly,
		}
	)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
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
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	assert.Nil(t, json.Unmarshal(bw.body.Bytes(), response))
	assert.True(t, response.Success)
	respData := response.Data.(map[string]interface{})
	respAvailableTimes := respData["availableTimes"].(float64)
	assert.Equal(t, availableTimes, int(respAvailableTimes))
}
