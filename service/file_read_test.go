//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestFileRead_Validate(t *testing.T) {

	var (
		fileReadSrv = &FileRead{
			Token: nil,
			File:  nil,
			IP:    nil,
		}

		err         error
		errValidate ValidateErrors
	)

	confirm := assert.New(t)
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	confirm.Nil(err)
	defer down(t)
	fileReadSrv.DB = trx

	errValidate = fileReadSrv.Validate()
	assert.NotNil(t, errValidate)
	confirm.NotNil(errValidate)
	confirm.True(errValidate.ContainsErrCode(10023))
	confirm.True(errValidate.ContainsErrCode(10024))
	confirm.Contains(errValidate.Error(), "invalid token")
	confirm.Contains(errValidate.Error(), "invalid file")

	token.Path = "/test"
	confirm.Nil(trx.Save(token).Error)

	dir, err := models.CreateOrGetLastDirectory(&token.App, "/save/to", trx)
	confirm.Nil(err)

	fileReadSrv.Token = token
	fileReadSrv.File = dir

	errValidate = fileReadSrv.Validate()
	confirm.NotNil(errValidate)
	confirm.Contains(errValidate.Error(), "file can't be accessed by some tokens")
}

func TestFileRead_Execute(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	token.AvailableTimes = 1000
	assert.Nil(t, trx.Save(token).Error)

	randomBytes := models.Random(556)
	randomBytesReader := bytes.NewReader(randomBytes)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	file, err := models.CreateFileFromReader(&token.App, "/test/random.bytes", randomBytesReader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, file.Object.Hash, randomBytesHash)

	fileReadSrv := &FileRead{
		BaseService: BaseService{
			DB:       trx,
			RootPath: &tempDir,
		},
		Token: token,
		File:  file,
	}

	assert.Nil(t, fileReadSrv.Validate())
	fileCreateValue, err := fileReadSrv.Execute(context.TODO())
	assert.Nil(t, err)
	fileCreateValueReader, ok := fileCreateValue.(io.Reader)
	assert.True(t, ok)
	allContent, err := ioutil.ReadAll(fileCreateValueReader)
	assert.Nil(t, err)
	allContentHash, err := util.Sha256Hash2String(allContent)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, allContentHash)

	assert.Nil(t, trx.Find(token).Error)
	assert.Equal(t, 999, token.AvailableTimes)
}
