//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"image"
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
	ctx.Set("inputParam", &ImageConvertInput{
		FileUID:       file.UID,
		OpenInBrowser: true,
		Type:          "zoom",
		Width:         200,
		Height:        200,
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

	input := ctx.MustGet("inputParam").(*ImageConvertInput)
	input.FileUID = ""

	ImageConvertHandler(ctx)
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

	ImageConvertHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the available times of token has already exhausted", response.Errors["ImageConvert.Token"][0])
}

func TestImageReadHandler3(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/image/convert")
		trx     *gorm.DB
		err     error
		down    func(*testing.T)
		token   *models.Token
		tempDir = models.NewTempDirForTest()
	)
	f, downImg := models.NewImageForTest(t)
	img, err := os.Open(f.Name())
	assert.Nil(t, err)
	testingChunkRootPath = &tempDir
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	testDBConn = trx
	defer func() {
		downImg(t)
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	file, err := models.CreateFileFromReader(&token.App, "/random.bytes", img, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?token=%s&fileUid=%s&width=%d&height=%d", api, token.UID, file.UID, 100, 200), nil)
	Routers().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, err)

	c, _, err := image.DecodeConfig(bytes.NewReader(w.Body.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, c.Width, 100)
	assert.Equal(t, c.Height, 200)
	resp := w.Result()
	assert.Equal(t, resp.Header.Get("Content-Type"), JpegContentType)

	var w2 = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?token=%s&fileUid=%s&type=%s&width=%d&height=%d", api, token.UID, file.UID, "thumb", 50, 100), nil)
	Routers().ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Nil(t, err)
	c, _, err = image.DecodeConfig(bytes.NewReader(w2.Body.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, c.Width, 50)
	assert.Equal(t, c.Height, 50)

	var w3 = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?token=%s&fileUid=%s&type=%s&width=%d&height=%d&left=%d&top=%d", api, token.UID, file.UID, "crop", 100, 100, 10, 10), nil)
	Routers().ServeHTTP(w3, req)

	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Nil(t, err)
	c, _, err = image.DecodeConfig(bytes.NewReader(w3.Body.Bytes()))
	assert.Nil(t, err)
	assert.Equal(t, c.Width, 100)
	assert.Equal(t, c.Height, 100)
}
