//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestFileCreate_Validate(t *testing.T) {
	trx, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	fileCreate := &FileCreate{
		BaseService: BaseService{
			DB: trx,
		},
		Token:     nil,
		Path:      strings.Repeat("1", 1001),
		Hidden:    2,
		Overwrite: 2,
		Rename:    2,
		Append:    2,
	}
	err := fileCreate.Validate()
	confirm := assert.New(t)
	confirm.NotNil(err)
	confirm.True(err.ContainsErrCode(10016))
	confirm.True(err.ContainsErrCode(10017))
	confirm.True(err.ContainsErrCode(10018))
	confirm.True(err.ContainsErrCode(10019))
	confirm.True(err.ContainsErrCode(10020))
	confirm.True(err.ContainsErrCode(10021))
	confirm.True(err.ContainsErrCode(10022))
	confirm.Contains(err.Error(), "path is not a legal unix path")
}

func TestFileCreate_Validate2(t *testing.T) {
	confirm := assert.New(t)
	expiredAt := time.Now().Add(10 * time.Hour)
	token, trx, down, err := models.NewTokenForTest(
		nil, t, "/test", &expiredAt, nil, nil, 10, 0)
	confirm.Nil(err)
	defer down(t)
	fileCreate := &FileCreate{
		BaseService: BaseService{
			DB: trx,
		},
		Token:     token,
		Path:      "/test/to",
		Hidden:    0,
		Overwrite: 0,
		Rename:    0,
		Append:    0,
	}
	err = fileCreate.Validate()
	confirm.Nil(err)
}

func newFileCreateForTest(t *testing.T, path string) (*FileCreate, func(*testing.T)) {
	token, trx, down, err := models.NewTokenForTest(nil, t, path, nil, nil, nil, 10, 0)
	tempDir := filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
	assert.Nil(t, err)
	fileCreate := &FileCreate{
		BaseService: BaseService{
			DB:       trx,
			RootPath: &tempDir,
		},
		Token: token,
	}
	return fileCreate, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}
}

// TestFileCreate_Execute is used to try to create a directory
func TestFileCreate_Execute(t *testing.T) {
	fileCreate, down := newFileCreateForTest(t, "/test")
	defer down(t)
	fileCreate.Path = "/create/dir"
	assert.Nil(t, fileCreate.Validate())

	fileValue, err := fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file, ok := fileValue.(*models.File)
	assert.True(t, ok)
	assert.NotNil(t, file)
	assert.True(t, file.ID > 0)
	assert.Equal(t, int8(1), file.IsDir)
	assert.Equal(t, file.App.ID, fileCreate.Token.App.ID)
	path, err := file.Path(fileCreate.DB)
	assert.Nil(t, err)
	assert.Equal(t, "/test/create/dir", path)

	fileValue, err = fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file2, ok := fileValue.(*models.File)
	assert.True(t, ok)
	assert.Equal(t, file2.ID, file.ID)
}

// TestFileCreate_Execute2 is used to test create file
func TestFileCreate_Execute2(t *testing.T) {
	fileCreate, down := newFileCreateForTest(t, "/test")
	defer down(t)
	randomBytes := models.Random(256)
	fileCreate.Reader = bytes.NewReader(randomBytes)
	fileCreate.Path = "/create/random.bytes"
	assert.Nil(t, fileCreate.Validate())

	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	fileValue, err := fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file, ok := fileValue.(*models.File)
	assert.True(t, ok)
	assert.Equal(t, 256, file.Size)
	assert.Equal(t, file.App.ID, fileCreate.Token.App.ID)
	assert.Equal(t, randomBytesHash, file.Object.Hash)
	assert.Equal(t, "bytes", file.Ext)
	path, err := file.Path(fileCreate.DB)
	assert.Nil(t, err)
	assert.Equal(t, "/test/create/random.bytes", path)
}

func newFileCreateForTestWithFile(t *testing.T) (*FileCreate, *models.File, hash.Hash, func(*testing.T)) {
	fileCreate, down := newFileCreateForTest(t, "/test")
	randomBytes := models.Random(256)
	fileCreate.Reader = bytes.NewReader(randomBytes)
	fileCreate.Path = "/create/random.bytes"
	assert.Nil(t, fileCreate.Validate())

	h := sha256.New()
	_, err := h.Write(randomBytes)
	assert.Nil(t, err)
	fileValue, err := fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file, ok := fileValue.(*models.File)
	assert.True(t, ok)
	return fileCreate, file, h, down
}

// TestFileCreate_Execute3 is used to overwrite file
func TestFileCreate_Execute3(t *testing.T) {
	fileCreate, file, h, down := newFileCreateForTestWithFile(t)
	defer down(t)
	assert.Equal(t, file.Object.Hash, hex.EncodeToString(h.Sum(nil)))

	randomBytes := models.Random(2*models.ChunkSize + 225)
	h2 := sha256.New()
	_, err := h2.Write(randomBytes)
	assert.Nil(t, err)
	fileCreate.Reader = bytes.NewReader(randomBytes)
	fileCreate.Overwrite = 1
	assert.Nil(t, fileCreate.Validate())

	fileValue, err := fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file2, ok := fileValue.(*models.File)
	assert.True(t, ok)
	assert.Equal(t, file2.ID, file.ID)
	assert.NotEqual(t, file2.ObjectID, file.ObjectID)
	assert.Equal(t, 2*models.ChunkSize+225, file2.Size)
	assert.Equal(t, file2.Object.Hash, hex.EncodeToString(h2.Sum(nil)))
	assert.Equal(t, 1, fileCreate.DB.Model(file2).Association("Histories").Count())

	path, err := file.Path(fileCreate.DB)
	assert.Nil(t, err)
	assert.Equal(t, "/test/create/random.bytes", path)
}

// TestFileCreate_Execute4 is used to append content to file
func TestFileCreate_Execute4(t *testing.T) {
	fileCreate, file, h, down := newFileCreateForTestWithFile(t)
	defer down(t)
	assert.Equal(t, file.Object.Hash, hex.EncodeToString(h.Sum(nil)))

	randomBytes := models.Random(2*models.ChunkSize + 225)
	_, err := h.Write(randomBytes)
	assert.Nil(t, err)
	fileCreate.Reader = bytes.NewReader(randomBytes)
	fileCreate.Append = 1
	assert.Nil(t, fileCreate.Validate())

	fileValue, err := fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file2, ok := fileValue.(*models.File)
	assert.True(t, ok)
	assert.Equal(t, file2.ID, file.ID)
	assert.Equal(t, file2.ObjectID, file.ObjectID)
	assert.Equal(t, file2.Size, file.Size+2*models.ChunkSize+225)
	assert.Equal(t, file2.Object.Hash, hex.EncodeToString(h.Sum(nil)))

	path, err := file.Path(fileCreate.DB)
	assert.Nil(t, err)
	assert.Equal(t, "/test/create/random.bytes", path)
}

// TestFileCreate_Execute5 try to rename file if the path has already existed
func TestFileCreate_Execute5(t *testing.T) {
	fileCreate, file, h, down := newFileCreateForTestWithFile(t)
	defer down(t)
	assert.Equal(t, file.Object.Hash, hex.EncodeToString(h.Sum(nil)))

	randomBytes := models.Random(2*models.ChunkSize + 225)
	h2 := sha256.New()
	_, err := h2.Write(randomBytes)
	assert.Nil(t, err)
	fileCreate.Reader = bytes.NewReader(randomBytes)
	fileCreate.Rename = 1
	assert.Nil(t, fileCreate.Validate())

	fileValue, err := fileCreate.Execute(context.TODO())
	assert.Nil(t, err)
	file2, ok := fileValue.(*models.File)
	assert.True(t, ok)
	assert.NotEqual(t, file2.ID, file.ID)
	assert.NotEqual(t, file2.ObjectID, file.ObjectID)
	assert.Equal(t, file2.Size, 2*models.ChunkSize+225)
	assert.Equal(t, file2.Object.Hash, hex.EncodeToString(h2.Sum(nil)))

	path, err := file2.Path(fileCreate.DB)
	assert.Nil(t, err)
	assert.NotEqual(t, "/test/create/random.bytes", path)
}

// TestFileCreate_Execute6 is used to test this case that the path has existed,
// if we continue to write content to this path, it should raise an error, because
// we don't set ant action like overwrite, rename or append
func TestFileCreate_Execute6(t *testing.T) {
	fileCreate, file, h, down := newFileCreateForTestWithFile(t)
	defer down(t)
	assert.Equal(t, file.Object.Hash, hex.EncodeToString(h.Sum(nil)))

	_, err := fileCreate.Execute(context.TODO())
	assert.NotNil(t, err)
	assert.Equal(t, ErrPathExisted, err)
}
