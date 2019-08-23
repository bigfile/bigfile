//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

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
		tempDir        = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
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
	fmt.Println(writer.body.String())
}
