//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/bigfile/bigfile/databases"

	"github.com/bigfile/bigfile/internal/util"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/stretchr/testify/assert"
)

func TestFileUpdate_Validate(t *testing.T) {
	var (
		hidden        int8 = 2
		path               = "/!!!/file"
		fileUpdateSrv      = &FileUpdate{
			Token:  nil,
			File:   nil,
			IP:     nil,
			Hidden: &hidden,
			Path:   &path,
		}

		err         error
		errValidate ValidateErrors
	)

	confirm := assert.New(t)
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	confirm.Nil(err)
	defer down(t)
	fileUpdateSrv.DB = trx

	errValidate = fileUpdateSrv.Validate()
	assert.NotNil(t, errValidate)
	confirm.NotNil(errValidate)
	confirm.True(errValidate.ContainsErrCode(10025))
	confirm.True(errValidate.ContainsErrCode(10026))
	confirm.True(errValidate.ContainsErrCode(10027))
	confirm.True(errValidate.ContainsErrCode(10028))
	confirm.Contains(errValidate.Error(), "invalid token")
	confirm.Contains(errValidate.Error(), "invalid file")

	token.Path = "/test"
	confirm.Nil(trx.Save(token).Error)

	dir, err := models.CreateOrGetLastDirectory(&token.App, "/save/to", trx)
	confirm.Nil(err)

	fileUpdateSrv.Token = token
	fileUpdateSrv.File = dir

	errValidate = fileUpdateSrv.Validate()
	confirm.NotNil(errValidate)
	confirm.Contains(errValidate.Error(), "file can't be accessed by some tokens")
}

func TestFileUpdate_Execute(t *testing.T) {
	tempDir := models.NewTempDirForTest()
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

	path := "/another/random1.bytes"
	var hidden int8 = 1
	fileUpdateSrv := &FileUpdate{
		BaseService: BaseService{
			DB:       trx,
			RootPath: &tempDir,
		},
		Token:  token,
		File:   file,
		Path:   &path,
		Hidden: &hidden,
	}

	assert.Nil(t, fileUpdateSrv.Validate())
	fileUpdateValue, err := fileUpdateSrv.Execute(context.TODO())
	assert.Nil(t, err)
	file, ok := fileUpdateValue.(*models.File)
	assert.True(t, ok)
	filePath, err := file.Path(trx)
	assert.Nil(t, err)
	assert.Equal(t, path, filePath)

	rootDir, err := models.CreateOrGetRootPath(&token.App, trx)
	assert.Nil(t, err)
	assert.Equal(t, 556, rootDir.Size)

	testDir, err := models.FindFileByPathWithTrashed(&token.App, "/test", trx)
	assert.Nil(t, err)
	assert.Equal(t, 0, testDir.Size)

	anotherDir, err := models.FindFileByPathWithTrashed(&token.App, "/another", trx)
	assert.Nil(t, err)
	assert.Equal(t, 556, anotherDir.Size)
}

func TestFileUpdate_Execute2(t *testing.T) {
	db := databases.MustNewConnection(nil)
	app, err := models.NewApp("TestFileCreate_Execute8", nil, db)
	assert.Nil(t, err)
	token, err := models.NewToken(app, "/", nil, nil, nil, 1000, 0, db)
	assert.Nil(t, err)
	tempDir := models.NewTempDirForTest()
	defer func() {
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	file, err := models.CreateFileFromReader(app, "/create/a/file.bytes", bytes.NewReader(models.Random(333)), models.Hidden, &tempDir, db)
	assert.Nil(t, err)
	assert.NotEqual(t, file.IsDir, models.IsDir)

	fileUpdateSrv := &FileUpdate{
		BaseService: BaseService{
			DB:       db,
			RootPath: &tempDir,
		},
		Token: token,
		File:  file,
	}
	fileUpdateVal, err := fileUpdateSrv.Execute(context.TODO())
	assert.Nil(t, err)
	fileUpdated, ok := fileUpdateVal.(*models.File)
	assert.True(t, ok)
	assert.Equal(t, file.ID, fileUpdated.ID)
}
