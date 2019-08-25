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

	"github.com/bigfile/bigfile/databases"

	"github.com/bigfile/bigfile/config"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func newFileUpdateForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
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
	ctx.Request, _ = http.NewRequest("PATCH", "http://bigfile.io", strings.NewReader(""))
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
	file, err := models.CreateFileFromReader(&token.App, "/test/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	path := "/another/another.bytes"
	ctx.Set("randomBytesHash", randomBytesHash)
	ctx.Set("path", path)
	ctx.Set("inputParam", &fileUpdateInput{
		FileUID: file.UID,
		Path:    &path,
	})

	return ctx, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

func TestFileUpdateHandler(t *testing.T) {
	ctx, down := newFileUpdateForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	input := ctx.MustGet("inputParam").(*fileUpdateInput)
	input.FileUID = ""

	FileUpdateHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "record not found", response.Errors["fileUid"][0])
}

func TestFileUpdateHandler2(t *testing.T) {
	ctx, down := newFileUpdateForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	db := ctx.MustGet("db").(*gorm.DB)
	token := ctx.MustGet("token").(*models.Token)
	token.AvailableTimes = 0
	assert.Nil(t, db.Save(token).Error)

	FileUpdateHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the available times of token has already exhausted", response.Errors["FileUpdate.Token"][0])
}

func TestFileUpdateHandler3(t *testing.T) {
	ctx, down := newFileUpdateForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)

	FileUpdateHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, ctx.GetString("randomBytesHash"), responseData["hash"].(string))
	assert.Equal(t, ctx.GetString("path"), responseData["path"].(string))
}

func TestFileUpdateHandler4(t *testing.T) {
	var (
		w       = httptest.NewRecorder()
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/update")
		trx     *gorm.DB
		err     error
		token   *models.Token
		down    func(*testing.T)
		tempDir = models.NewTempDirForTest()
	)

	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	testingChunkRootPath = &tempDir
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	testDBConn = trx

	randomBytes := models.Random(128)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	randomBytesReader := bytes.NewReader(randomBytes)
	file, err := models.CreateFileFromReader(&token.App, "/test/random.bytes", randomBytesReader, int8(0), testingChunkRootPath, trx)
	assert.Nil(t, err)

	body := fmt.Sprintf("token=%s&fileUid=%s&path=/another/another.bytes&nonce=%s", token.UID, file.UID, models.RandomWithMd5(2))
	req, _ := http.NewRequest("PATCH", api, strings.NewReader(body))
	req.Header.Set("X-Forwarded-For", "192.168.0.1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	Routers().ServeHTTP(w, req)
	response, err := parseResponse(w.Body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, randomBytesHash, responseData["hash"].(string))
	assert.Equal(t, "/another/another.bytes", responseData["path"].(string))
}

func BenchmarkFileUpdateHandler(b *testing.B) {
	b.StopTimer()

	var (
		api     = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/update")
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

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		body := fmt.Sprintf("token=%s&fileUid=%s&path=/another/another.bytes&nonce=%s", token.UID, file.UID, models.RandomWithMd5(2))
		req, _ := http.NewRequest("PATCH", api, strings.NewReader(body))
		req.Header.Set("X-Forwarded-For", "192.168.0.1")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		Routers().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatal("response code should be 200")
		}
	}
}
