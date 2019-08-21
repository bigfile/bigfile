//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo/bson"
)

func TestFile_TableName(t *testing.T) {
	assert.Equal(t, (&File{}).TableName(), "files")
}

func TestCreateOrGetRootPath(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	file, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, app.ID, file.AppID)
}

func TestCreateOrGetLastDirectory(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	file, err := CreateOrGetLastDirectory(app, "/save/to/images", trx)
	assert.Nil(t, err)
	assert.Equal(t, app.ID, file.AppID)
	assert.Equal(t, int8(1), file.IsDir)
	assert.Equal(t, "images", file.Name)
	var subDirCount int
	assert.Nil(t, trx.Model(&File{}).Where("appId = ?", app.ID).Count(&subDirCount).Error)
	assert.Equal(t, 4, subDirCount)
}

func TestFile_UpdateParentSize(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	file, err := CreateOrGetLastDirectory(app, "/save/to/images", trx)
	assert.Nil(t, err)

	assert.Nil(t, file.UpdateParentSize(1000, trx))
	root, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 1000, root.Size)

	assert.Nil(t, file.UpdateParentSize(-100, trx))
	root, err = CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 900, root.Size)
}

func TestFindFileByPath(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	_, err = CreateOrGetLastDirectory(app, "/save/to/images", trx)
	assert.Nil(t, err)

	imagesDir, err := FindFileByPath(app, "/save/to/images", trx)
	assert.Nil(t, err)
	assert.Equal(t, int8(1), imagesDir.IsDir)
	assert.Equal(t, app.ID, imagesDir.AppID)
	assert.Equal(t, "images", imagesDir.Name)

	assert.Nil(t, trx.Save(&File{
		UID:      bson.NewObjectId().Hex(),
		PID:      imagesDir.ID,
		AppID:    app.ID,
		ObjectID: 0,
		Size:     12,
		Name:     "test.png",
		Ext:      "png",
		IsDir:    0,
	}).Error)

	testPngFile, err := FindFileByPath(app, "/save/to/images/test.png", trx)
	assert.Nil(t, err)
	assert.Equal(t, int8(0), testPngFile.IsDir)
	assert.Equal(t, app.ID, testPngFile.AppID)
	assert.Equal(t, "test.png", testPngFile.Name)
	assert.Equal(t, imagesDir.ID, testPngFile.PID)

	_, err = FindFileByPath(app, "/save/to/images/test.jpg", trx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestCreateFileFromReader(t *testing.T) {
	var (
		app             *App
		trx             *gorm.DB
		err             error
		file            *File
		down            func(*testing.T)
		randomBytes     = Random(uint(ChunkSize*2 + 145))
		randomBytesHash string
		reader          = bytes.NewReader(randomBytes)
		tempDir         = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
	)

	app, trx, down, err = newAppForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	file, err = CreateFileFromReader(app, "/save/to/random.txt", reader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, ChunkSize*2+145, file.Size)
	assert.Equal(t, "random.txt", file.Name)
	assert.Equal(t, ChunkSize*2+145, file.Object.Size)
	assert.Equal(t, randomBytesHash, file.Object.Hash)

	root, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, ChunkSize*2+145, root.Size)

	_, err = CreateFileFromReader(app, "/save/to/random.txt", strings.NewReader(""), int8(0), &tempDir, trx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ErrFileExisted.Error())
}

func TestFile_AppendFromReader(t *testing.T) {
	var (
		h           = sha256.New()
		app         *App
		trx         *gorm.DB
		err         error
		file        *File
		down        func(*testing.T)
		randomBytes = Random(uint(ChunkSize*2 + 145))
		reader      = bytes.NewReader(randomBytes)
		tempDir     = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
	)

	app, trx, down, err = newAppForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	_, err = h.Write(randomBytes)
	assert.Nil(t, err)
	assert.Nil(t, err)
	file, err = CreateFileFromReader(app, "/save/to/random.txt", reader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), file.Object.Hash)

	randomBytes = Random(uint(256))
	_, err = h.Write(randomBytes)
	assert.Nil(t, err)
	assert.Nil(t, file.AppendFromReader(bytes.NewBuffer(randomBytes), int8(0), &tempDir, trx))
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), file.Object.Hash)
	assert.Equal(t, ChunkSize*2+145+256, file.Size)
	assert.Equal(t, ChunkSize*2+145+256, file.Object.Size)

	root, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, ChunkSize*2+145+256, root.Size)
}
