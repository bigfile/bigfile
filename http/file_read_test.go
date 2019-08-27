//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newFileReadForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
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

	randomBytes := models.Random(128)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	randomBytesReader := bytes.NewReader(randomBytes)
	file, err := models.CreateFileFromReader(&token.App, "/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	ctx.Set("randomBytesHash", randomBytesHash)
	ctx.Set("inputParam", &fileReadInput{
		FileUID: file.UID,
	})

	return ctx, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

func TestFileReadHandler(t *testing.T) {
	ctx, down := newFileReadForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	input := ctx.MustGet("inputParam").(*fileReadInput)
	input.FileUID = ""

	FileReadHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "record not found", response.Errors["fileUid"][0])
}

func TestFileReadHandler2(t *testing.T) {
	ctx, down := newFileReadForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.AvailableTimes = 0
	assert.Nil(t, db.Save(token).Error)

	FileReadHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the available times of token has already exhausted", response.Errors["FileRead.Token"][0])
}

func TestFileReadHandler3(t *testing.T) {
	ctx, down := newFileReadForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)
	FileReadHandler(ctx)

	bodyHash, err := util.Sha256Hash2String(writer.body.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, bodyHash, ctx.GetString("randomBytesHash"))
}

func TestFileReadHandler4(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/read")
		trx     *gorm.DB
		err     error
		down    func(*testing.T)
		token   *models.Token
		tempDir = models.NewTempDirForTest()
	)

	testingChunkRootPath = &tempDir
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	testDBConn = trx
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	randomBytes := models.Random(128)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	randomBytesReader := bytes.NewReader(randomBytes)
	file, err := models.CreateFileFromReader(&token.App, "/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?token=%s&fileUid=%s", api, token.UID, file.UID), nil)
	Routers().ServeHTTP(w, req)
	responseBodyHash, err := util.Sha256Hash2String(w.Body.Bytes())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, err)
	assert.Equal(t, responseBodyHash, randomBytesHash)
}

func TestFileReadHandler5(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/read")
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
	assert.Nil(t, trx.Save(token).Error)
	testDBConn = trx
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	randomBytes := models.Random(128)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	randomBytesReader := bytes.NewReader(randomBytes)
	file, err := models.CreateFileFromReader(&token.App, "/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	qs := getParamsSignBody(map[string]interface{}{
		"token":   token.UID,
		"fileUid": file.UID,
		"nonce":   models.RandomWithMd5(333),
	}, secret)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
	Routers().ServeHTTP(w, req)
	responseBodyHash, err := util.Sha256Hash2String(w.Body.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, responseBodyHash, randomBytesHash)
}

func BenchmarkFileReadHandler(b *testing.B) {
	b.StopTimer()

	var (
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/read")
		trx     *gorm.DB
		err     error
		token   *models.Token
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

	token, err = models.NewToken(app, "/benchmark", nil, nil, nil, 100, 0, trx)
	if err != nil {
		b.Fatal(err)
	}

	randomBytes := models.Random(128)
	randomBytesReader := bytes.NewReader(randomBytes)
	file, err := models.CreateFileFromReader(&token.App, "/benchmark/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	if err != nil {
		b.Fatal(err)
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?token=%s&fileUid=%s", api, token.UID, file.UID), nil)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		Routers().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatal("response code should be 200")
		}
	}
}
