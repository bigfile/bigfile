//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func newFileCreateForTest(t *testing.T) (*gin.Context, func(*testing.T)) {
	var (
		ctx            *gin.Context
		trx            *gorm.DB
		err            error
		token          *models.Token
		down           func(*testing.T)
		tempDir        = models.NewTempDirForTest()
		body           = &bytes.Buffer{}
		formBodyWriter = multipart.NewWriter(body)
	)

	testingChunkRootPath = &tempDir
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Writer = &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
	assert.Nil(t, formBodyWriter.WriteField("token", token.UID))
	assert.Nil(t, formBodyWriter.Close())
	ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", body)
	ctx.Request.Header.Set("X-Forwarded-For", "192.168.0.1")
	ctx.Request.Header.Set("Content-Type", formBodyWriter.FormDataContentType())
	ctx.Set("db", trx)
	ctx.Set("token", token)
	ctx.Set("inputParam", &fileCreateInput{
		Path: "/create/a/directory",
	})
	reqRecord := models.MustNewRequestWithProtocol("http", trx)
	ctx.Set("reqRecord", reqRecord)
	ctx.Set("requestId", int64(reqRecord.ID))
	return ctx, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

// TestFileCreateHandler is only used to FileCreateHandler, no any
// middleware participate
func TestFileCreateHandler(t *testing.T) {
	ctx, down := newFileCreateForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)
	FileCreateHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, 1, int(responseData["isDir"].(float64)))
	assert.Equal(t, 0, int(responseData["size"].(float64)))
	assert.Equal(t, 0, int(responseData["hidden"].(float64)))
	assert.Equal(t, "/create/a/directory", responseData["path"].(string))
}

func TestFileCreateHandler4(t *testing.T) {
	ctx, down := newFileCreateForTest(t)
	defer down(t)
	writer := ctx.Writer.(*bodyWriter)
	input := ctx.MustGet("inputParam").(*fileCreateInput)
	hidden := true
	input.Hidden = &hidden
	ctx.Set("token", (*models.Token)(nil))
	FileCreateHandler(ctx)
	response, err := parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, writer.body.String(), "invalid token")
}

// TestFileCreateHandler2 is used to test upload file
func TestFileCreateHandler2(t *testing.T) {
	var (
		err             error
		ctx             *gin.Context
		size            int
		down            func(*testing.T)
		writer          *bodyWriter
		response        *Response
		trueValue       = true
		falseValue      = false
		responseData    map[string]interface{}
		randomBytesHash string
	)

	ctx, down = newFileCreateForTest(t)
	defer down(t)
	writer = ctx.Writer.(*bodyWriter)

	input := ctx.MustGet("inputParam").(*fileCreateInput)
	input.Path = "/save/to/random.bytes"

	setRequestForCtx := func(ctx *gin.Context, size int) string {
		var (
			body           = &bytes.Buffer{}
			formBodyWriter = multipart.NewWriter(body)
			formFileWriter io.Writer
		)
		formFileWriter, err = formBodyWriter.CreateFormFile("file", "random.bytes")
		assert.Nil(t, err)
		randomBytes := models.Random(uint(size))
		randomBytesHash, err := util.Sha256Hash2String(randomBytes)
		assert.Nil(t, err)
		_, err = formFileWriter.Write(randomBytes)
		assert.Nil(t, err)
		assert.Nil(t, formBodyWriter.Close())
		ctx.Request, _ = http.NewRequest("POST", "http://bigfile.io", body)
		ctx.Request.Header.Set("X-Forwarded-For", "192.168.0.1")
		ctx.Request.Header.Set("Content-Type", formBodyWriter.FormDataContentType())
		return randomBytesHash
	}

	// chunk size exceed limit
	randomBytesHash = setRequestForCtx(ctx, models.ChunkSize+1)
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, models.ErrChunkExceedLimit.Error(), response.Errors["file"][0])
	writer.body.Reset()

	// everything is ok
	randomBytesHash = setRequestForCtx(ctx, 256)
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData = response.Data.(map[string]interface{})
	assert.Equal(t, 0, int(responseData["isDir"].(float64)))
	assert.Equal(t, 256, int(responseData["size"].(float64)))
	assert.Equal(t, 0, int(responseData["hidden"].(float64)))
	assert.Equal(t, "/save/to/random.bytes", responseData["path"].(string))
	assert.Equal(t, randomBytesHash, responseData["hash"].(string))
	writer.body.Reset()

	// size error
	input.Path = "/save/to/random1.bytes"
	size = 255
	input.Size = &size
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the size of file doesn't match", response.Errors["size"][0])
	writer.body.Reset()

	// hash error
	input.Path = "/save/to/random2.bytes"
	fakeHash := "fake hash"
	size = 256
	input.Hash = &fakeHash
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the hash of file doesn't match", response.Errors["hash"][0])
	writer.body.Reset()

	// hash and size both are right
	size = 278
	randomBytesHash = setRequestForCtx(ctx, size)
	input.Path = "/save/to/random3.bytes"
	input.Hash = &randomBytesHash
	input.Size = &size
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData = response.Data.(map[string]interface{})
	assert.Equal(t, "/save/to/random3.bytes", responseData["path"].(string))
	assert.Equal(t, size, int(responseData["size"].(float64)))
	writer.body.Reset()

	// path has been occupied
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "the path has already existed", response.Errors["system"][0])
	writer.body.Reset()

	// path has been occupied, but append
	input.Append = &trueValue
	input.Rename = &falseValue
	input.Overwrite = &falseValue
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData = response.Data.(map[string]interface{})
	assert.Equal(t, size*2, int(responseData["size"].(float64)))
	assert.Equal(t, "/save/to/random3.bytes", responseData["path"].(string))
	writer.body.Reset()

	// path has been occupied, but Overwrite
	size = 222
	randomBytesHash = setRequestForCtx(ctx, size)
	input.Path = "/save/to/random3.bytes"
	input.Hash = &randomBytesHash
	input.Size = &size
	input.Append = &falseValue
	input.Rename = &falseValue
	input.Overwrite = &trueValue
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData = response.Data.(map[string]interface{})
	assert.Equal(t, size, int(responseData["size"].(float64)))
	assert.Equal(t, randomBytesHash, responseData["hash"].(string))
	assert.Equal(t, "/save/to/random3.bytes", responseData["path"].(string))
	writer.body.Reset()

	// path has been occupied, but rename
	size = 234
	randomBytesHash = setRequestForCtx(ctx, size)
	input.Path = "/save/to/random3.bytes"
	input.Hash = &randomBytesHash
	input.Size = &size
	input.Append = &falseValue
	input.Rename = &trueValue
	input.Overwrite = &falseValue
	FileCreateHandler(ctx)
	response, err = parseResponse(writer.body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData = response.Data.(map[string]interface{})
	assert.Equal(t, size, int(responseData["size"].(float64)))
	assert.Equal(t, randomBytesHash, responseData["hash"].(string))
	assert.NotEqual(t, "/save/to/random3.bytes", responseData["path"].(string))
	writer.body.Reset()
}

func TestFileCreateHandler3(t *testing.T) {
	var (
		w              = httptest.NewRecorder()
		api            = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/create")
		trx            *gorm.DB
		err            error
		down           func(*testing.T)
		body           = &bytes.Buffer{}
		token          *models.Token
		secret         = models.RandomWithMD5(122)
		params         map[string]interface{}
		tempDir        = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		response       *Response
		formBodyWriter = multipart.NewWriter(body)
		formFileWriter io.Writer
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

	token.Secret = &secret
	assert.Nil(t, trx.Save(token).Error)

	params = map[string]interface{}{
		"token": token.UID,
		"path":  "/test/create/binary/file.binary",
		"nonce": models.RandomWithMD5(255),
	}
	params["sign"] = getParamsSignature(params, secret)
	for k, v := range params {
		assert.Nil(t, formBodyWriter.WriteField(k, v.(string)))
	}

	formFileWriter, err = formBodyWriter.CreateFormFile("file", "random.bytes")
	assert.Nil(t, err)
	randomBytes := models.Random(uint(399))
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	_, err = formFileWriter.Write(randomBytes)
	assert.Nil(t, err)
	assert.Nil(t, formBodyWriter.Close())

	req, _ := http.NewRequest("POST", api, body)
	req.Header.Set("X-Forwarded-For", "192.168.0.1")
	req.Header.Set("Content-Type", formBodyWriter.FormDataContentType())
	Routers().ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	response, err = parseResponse(w.Body.String())
	assert.Nil(t, err)
	assert.True(t, response.Success)
	responseData := response.Data.(map[string]interface{})
	assert.Equal(t, 399, int(responseData["size"].(float64)))
	assert.Equal(t, randomBytesHash, responseData["hash"].(string))
	assert.Equal(t, "/test/create/binary/file.binary", responseData["path"].(string))
}

func BenchmarkFileCreateHandler(b *testing.B) {
	b.StopTimer()
	var (
		api            = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/file/create")
		trx            *gorm.DB
		err            error
		body           = &bytes.Buffer{}
		token          *models.Token
		secret         = models.RandomWithMD5(122)
		params         map[string]interface{}
		tempDir        = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		formBodyWriter = multipart.NewWriter(body)
		formFileWriter io.Writer
	)

	trx = databases.MustNewConnection(nil).BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
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

	token, err = models.NewToken(app, "/benchmark", nil, nil, &secret, 100, 0, trx)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		params = map[string]interface{}{
			"token":  token.UID,
			"path":   "/test/create/binary/file.binary",
			"nonce":  models.RandomWithMD5(255),
			"append": "1",
		}
		params["sign"] = getParamsSignature(params, secret)
		for k, v := range params {
			if err = formBodyWriter.WriteField(k, v.(string)); err != nil {
				b.Fatal(err)
			}
		}
		if formFileWriter, err = formBodyWriter.CreateFormFile("file", "random.bytes"); err != nil {
			b.Fatal(err)
		}
		if _, err = formFileWriter.Write(models.Random(uint(399))); err != nil {
			b.Fatal(err)
		}
		if err = formBodyWriter.Close(); err != nil {
			b.Fatal(err)
		}
		req, _ := http.NewRequest("POST", api, body)
		req.Header.Set("X-Forwarded-For", "192.168.0.1")
		req.Header.Set("Content-Type", formBodyWriter.FormDataContentType())
		Routers().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatal("response code should be 200")
		}
	}
}
