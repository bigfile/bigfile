//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

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

func TestFileDelete_Validate(t *testing.T) {
	var (
		fileDeleteSrv = &FileDelete{
			Token: nil,
			File:  nil,
		}

		err         error
		errValidate ValidateErrors
	)

	confirm := assert.New(t)
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	confirm.Nil(err)
	defer down(t)
	fileDeleteSrv.DB = trx

	errValidate = fileDeleteSrv.Validate()
	assert.NotNil(t, errValidate)
	confirm.NotNil(errValidate)
	confirm.True(errValidate.ContainsErrCode(10029))
	confirm.True(errValidate.ContainsErrCode(10030))
	confirm.Contains(errValidate.Error(), "invalid token")
	confirm.Contains(errValidate.Error(), "invalid file")

	token.Path = "/test"
	confirm.Nil(trx.Save(token).Error)

	dir, err := models.CreateOrGetLastDirectory(&token.App, "/save/to", trx)
	confirm.Nil(err)

	fileDeleteSrv.Token = token
	fileDeleteSrv.File = dir

	errValidate = fileDeleteSrv.Validate()
	confirm.NotNil(errValidate)
	confirm.Contains(errValidate.Error(), "file can't be accessed by some tokens")
}

func TestFileDelete_Execute(t *testing.T) {
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
	file, err := models.CreateFileFromReader(&token.App, "/test/to/bytes/random.bytes", randomBytesReader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, file.Object.Hash, randomBytesHash)

	rootDir, err := models.CreateOrGetRootPath(&token.App, trx)
	assert.Nil(t, err)
	assert.Equal(t, 556, rootDir.Size)

	fileDeleteSrv := &FileDelete{
		BaseService: BaseService{DB: trx},
		Token:       token,
		File:        file,
	}
	assert.Nil(t, fileDeleteSrv.Validate())

	fileDeleteValue, err := fileDeleteSrv.Execute(context.TODO())
	assert.Nil(t, err)
	file, ok := fileDeleteValue.(*models.File)
	assert.True(t, ok)
	assert.NotNil(t, file.DeletedAt)

	rootDir, err = models.CreateOrGetRootPath(&token.App, trx)
	assert.Nil(t, err)
	assert.Equal(t, 0, rootDir.Size)
}

func TestFileDelete_Execute2(t *testing.T) {
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
	file, err := models.CreateFileFromReader(&token.App, "/test/to/bytes/random.bytes", randomBytesReader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, file.Object.Hash, randomBytesHash)

	rootDir, err := models.CreateOrGetRootPath(&token.App, trx)
	assert.Nil(t, err)
	assert.Equal(t, 556, rootDir.Size)

	fileDeleteSrv := &FileDelete{
		BaseService: BaseService{DB: trx},
		Token:       token,
		File:        file.Parent,
	}
	assert.Nil(t, fileDeleteSrv.Validate())

	_, err = fileDeleteSrv.Execute(context.TODO())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "delete non-empty directory")

	toDir, err := models.FindFileByPathWithTrashed(&token.App, "/test/to", trx)
	assert.Nil(t, err)
	assert.Equal(t, 556, toDir.Size)

	trueValue := true
	fileDeleteSrv.Force = &trueValue
	_, err = fileDeleteSrv.Execute(context.TODO())
	assert.Nil(t, err)

	rootDir, err = models.CreateOrGetRootPath(&token.App, trx)
	assert.Nil(t, err)
	assert.Equal(t, 0, rootDir.Size)
}

func TestFileDelete_Execute3(t *testing.T) {
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

	fileDelete := &FileDelete{
		BaseService: BaseService{DB: db, RootPath: &tempDir},
		Token:       token,
		File:        file,
	}
	fileDeleteVal, err := fileDelete.Execute(context.TODO())
	assert.Nil(t, err)
	fileDeleted, ok := fileDeleteVal.(*models.File)
	assert.True(t, ok)
	assert.NotNil(t, fileDeleted.DeletedAt)
}
