//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/databases"

	"github.com/bigfile/bigfile/config"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newDirectoryListForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
	var (
		ctx     *gin.Context
		trx     *gorm.DB
		err     error
		token   *models.Token
		down    func(*testing.T)
		tempDir = models.NewTempDirForTest()
	)

	testingChunkRootPath = &tempDir
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Writer = &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", strings.NewReader(""))
	ctx.Request.Header.Set("X-Forwarded-For", "192.168.0.1")
	ctx.Set("db", trx)
	ctx.Set("token", token)
	reqRecord := models.MustNewRequestWithProtocol("http", trx)
	ctx.Set("reqRecord", reqRecord)
	ctx.Set("requestId", int64(reqRecord.ID))

	for index := 0; index < 18; index++ {
		_, err = models.CreateOrGetLastDirectory(&token.App, "/test/directory/list/"+strconv.Itoa(index), trx)
		assert.Nil(t, err)
	}

	subDir := "/"
	sort := "-type"
	limit := 10
	offset := 0
	ctx.Set("inputParam", &directoryListInput{
		SubDir: &subDir,
		Sort:   &sort,
		Limit:  &limit,
		Offset: &offset,
	})

	return ctx, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

func TestDirectoryListHandler(t *testing.T) {
	ctx, down := newDirectoryListForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.Path = "/not/exist"
	assert.Nil(t, db.Model(token).Update("path", token.Path).Error)

	DirectoryListHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "record not found", response.Errors["system"][0])
}

func TestDirectoryListHandler2(t *testing.T) {
	ctx, down := newDirectoryListForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	DirectoryListHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, int(responseData["total"].(float64)))
	assert.Equal(t, 1, int(responseData["pages"].(float64)))
	assert.Equal(t, 1, len(responseData["items"].([]interface{})))
}

func TestDirectoryListHandler3(t *testing.T) {
	ctx, down := newDirectoryListForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.Path = "/test/directory/list"
	assert.Nil(t, db.Model(token).Update("path", token.Path).Error)

	DirectoryListHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 18, int(responseData["total"].(float64)))
	assert.Equal(t, 2, int(responseData["pages"].(float64)))
	assert.Equal(t, 10, len(responseData["items"].([]interface{})))
}

func TestDirectoryListHandler4(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/directory/list")
		trx     *gorm.DB
		err     error
		down    func(*testing.T)
		token   *models.Token
		secret  = models.RandomWithMd5(222)
		tempDir = models.NewTempDirForTest()
	)

	testingChunkRootPath = &tempDir
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	token.Secret = &secret
	token.Path = "/test/directory"
	assert.Nil(t, trx.Save(token).Error)
	testDBConn = trx
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	for index := 0; index < 18; index++ {
		_, err = models.CreateOrGetLastDirectory(&token.App, "/test/directory/list/"+strconv.Itoa(index), trx)
		assert.Nil(t, err)
	}

	qs := getParamsSignBody(map[string]interface{}{
		"token":  token.UID,
		"nonce":  models.RandomWithMd5(222),
		"subDir": "/list/",
	}, secret)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
	Routers().ServeHTTP(w, req)
	response, err := parseResponse(w.Body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 18, int(responseData["total"].(float64)))
	assert.Equal(t, 2, int(responseData["pages"].(float64)))
	assert.Equal(t, 10, len(responseData["items"].([]interface{})))
}

func BenchmarkDirectoryListHandler(b *testing.B) {
	b.StopTimer()

	var (
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/directory/list")
		trx     *gorm.DB
		err     error
		token   *models.Token
		secret  = models.RandomWithMd5(222)
		tempDir = models.NewTempDirForTest()
	)

	trx = databases.MustNewConnection(nil).Begin()
	testDBConn = trx
	testingChunkRootPath = &tempDir
	defer func() {
		trx.Rollback()
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	note := "test"
	app, err := models.NewApp("test", &note, trx)
	if err != nil {
		b.Fatal(err)
	}

	token, err = models.NewToken(app, "/test/directory", nil, nil, &secret, 100, 0, trx)
	if err != nil {
		b.Fatal(err)
	}

	for index := 0; index < 18; index++ {
		if _, err = models.CreateOrGetLastDirectory(&token.App, "/test/directory/list/"+strconv.Itoa(index), trx); err != nil {
			b.Fatal(err)
		}
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		qs := getParamsSignBody(map[string]interface{}{
			"token":  token.UID,
			"nonce":  models.RandomWithMd5(222),
			"subDir": "/list/",
		}, secret)
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
		Routers().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatal("response code should be 200")
		}
	}
}
