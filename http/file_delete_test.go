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
	"strings"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newFileDeleteForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
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
	file, err := models.CreateFileFromReader(&token.App, "/save/to/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)
	assert.Equal(t, file.Object.Hash, randomBytesHash)
	filePath, err := file.Path(trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/random.bytes", filePath)

	ctx.Set("inputParam", &fileDeleteInput{
		FileUID: file.UID,
		Force:   true,
	})

	return ctx, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

func TestFileDeleteHandler(t *testing.T) {
	ctx, down := newFileDeleteForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	input := ctx.MustGet("inputParam").(*fileDeleteInput)
	input.FileUID = ""

	FileDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "record not found", response.Errors["fileUid"][0])
}

func TestFileDeleteHandler2(t *testing.T) {
	ctx, down := newFileDeleteForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.AvailableTimes = 0
	assert.Nil(t, db.Save(token).Error)

	FileDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the available times of token has already exhausted", response.Errors["FileDelete.Token"][0])
}

func TestFileDeleteHandler3(t *testing.T) {
	ctx, down := newFileDeleteForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.Path = "/test/any/dir"
	assert.Nil(t, db.Save(token).Error)

	FileDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "file can't be accessed by some tokens", response.Errors["FileDelete.Token"][0])
}

func TestFileDeleteHandler4(t *testing.T) {
	ctx, down := newFileDeleteForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	FileDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, responseData["deletedAt"])
}

func TestFileDeleteHandler5(t *testing.T) {
	ctx, down := newFileDeleteForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)
	token := ctx.MustGet("token").(*models.Token)

	db := ctx.MustGet("db").(*gorm.DB)
	toDir, err := models.FindFileByPathWithTrashed(&token.App, "/save/to", db)
	assert.Nil(t, err)

	input := ctx.MustGet("inputParam").(*fileDeleteInput)
	input.FileUID = toDir.UID
	input.Force = true

	FileDeleteHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)

	rootDir, err := models.CreateOrGetRootPath(&token.App, db)
	assert.Nil(t, err)
	assert.Equal(t, 0, rootDir.Size)

	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, responseData["deletedAt"])
}

func TestFileDeleteHandler6(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/delete")
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
	assert.Nil(t, err)
	randomBytesReader := bytes.NewReader(randomBytes)
	file, err := models.CreateFileFromReader(&token.App, "/save/to/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	apiWithQs := fmt.Sprintf("%s?token=%s&fileUid=%s&nonce=%s", api, token.UID, file.UID, models.RandomWithMd5(222))
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
