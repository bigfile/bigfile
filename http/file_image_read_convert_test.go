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
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newImageConvertForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
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
	file, err := models.CreateFileFromReader(&token.App, "/random.png", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	ctx.Set("randomBytesHash", randomBytesHash)
	ctx.Set("inputParam", &imageFileReadInput{
		FileUID:       file.UID,
		OpenInBrowser: true,
	})

	return ctx, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

func TestImageReadHandler(t *testing.T) {
	ctx, down := newImageConvertForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	input := ctx.MustGet("inputParam").(*imageFileReadInput)
	input.FileUID = ""

	ImageFileConvertHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "record not found", response.Errors["fileUid"][0])
}

func TestImageReadHandler2(t *testing.T) {
	ctx, down := newImageConvertForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.AvailableTimes = 0
	assert.Nil(t, db.Save(token).Error)

	ImageFileConvertHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the available times of token has already exhausted", response.Errors["FileRead.Token"][0])
}

func TestImageReadHandler4(t *testing.T) {
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

func TestImageReadHandler5(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/read")
		trx     *gorm.DB
		err     error
		down    func(*testing.T)
		token   *models.Token
		secret  = models.RandomWithMD5(222)
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
		"nonce":   models.RandomWithMD5(333),
	}, secret)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
	Routers().ServeHTTP(w, req)
	responseBodyHash, err := util.Sha256Hash2String(w.Body.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, responseBodyHash, randomBytesHash)
}

func TestImageReadHandler6(t *testing.T) {
	ctx, down := newImageConvertForTest(t)
	defer down(t)

	writer := ctx.Writer.(*bodyWriter)
	input := ctx.MustGet("inputParam").(*imageFileReadInput)
	db := ctx.MustGet("db").(*gorm.DB)

	file, err := models.FindFileByUID(input.FileUID, false, db)
	assert.Nil(t, err)
	assert.Nil(t, db.Preload("Parent").First(file).Error)
	assert.NotNil(t, file.Parent)
	input.FileUID = file.Parent.UID

	ImageFileConvertHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, models.ErrReadDir.Error(), response.Errors["system"][0])
}
