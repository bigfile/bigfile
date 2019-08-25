//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
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
	assert.Equal(t, app.ID, file.App.ID)
	file.Size = 255
	assert.Nil(t, trx.Save(file).Error)

	rootDir, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, rootDir.Size, file.Size)
}

func TestCreateOrGetLastDirectory(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	images, err := CreateOrGetLastDirectory(app, "/save/to/images", trx)
	assert.Nil(t, err)
	assert.Equal(t, app.ID, images.AppID)
	assert.Equal(t, app.ID, images.App.ID)
	assert.Equal(t, int8(1), images.IsDir)
	assert.Equal(t, "images", images.Name)
	images.Size = 255
	assert.Nil(t, trx.Save(images).Error)
	var subDirCount int
	assert.Nil(t, trx.Model(&File{}).Where("appId = ?", app.ID).Count(&subDirCount).Error)
	assert.Equal(t, 4, subDirCount)

	images2, err := CreateOrGetLastDirectory(app, "/save/to/images", trx)
	assert.Nil(t, err)
	assert.Equal(t, images.ID, images2.ID)
	assert.Equal(t, images.Size, images2.Size)
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
	assert.Equal(t, app.ID, testPngFile.App.ID)
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
	assert.Equal(t, app.ID, file.App.ID)
	assert.Equal(t, app.ID, file.AppID)
	assert.Equal(t, "txt", file.Ext)

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
	assert.Equal(t, app.ID, file.App.ID)
	assert.Equal(t, app.ID, file.AppID)

	root, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, ChunkSize*2+145+256, root.Size)
}

func TestFile_Path(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	dir, err := CreateOrGetLastDirectory(app, "/save/to/images", trx)
	assert.Nil(t, err)
	path, err := dir.Path(trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/images", path)

	file := &File{UID: bson.NewObjectId().Hex(), PID: dir.ID, Name: "test.png", AppID: app.ID}
	assert.Nil(t, trx.Save(file).Error)
	path, err = file.Path(trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/images/test.png", path)
}

func TestFile_OverWriteFromReader(t *testing.T) {
	var (
		trx             *gorm.DB
		app             *App
		err             error
		down            func(*testing.T)
		file            *File
		object          Object
		tempDir         = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		randomBytes     = Random(uint(127))
		randomBytesHash string
		reader          = bytes.NewReader(randomBytes)
	)
	app, trx, down, err = newAppForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	file, err = CreateFileFromReader(app, "/test/random/content/txt.byte", reader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, file.Object.Hash)
	object = file.Object

	root, err := CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 127, root.Size)

	randomBytes = Random(uint(120))
	reader = bytes.NewReader(randomBytes)
	assert.Nil(t, file.OverWriteFromReader(reader, int8(0), &tempDir, trx))
	assert.NotEqual(t, file.Object.ID, object.ID)
	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, file.Object.Hash)
	assert.Equal(t, app.ID, file.App.ID)
	assert.Equal(t, app.ID, file.AppID)
	assert.Equal(t, 120, file.Size)
	assert.Equal(t, 1, trx.Model(file).Association("Histories").Count())

	root, err = CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 120, root.Size)
}

func TestFile_Reader(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	file, err := CreateOrGetLastDirectory(app, "/test", trx)
	assert.Nil(t, err)
	_, err = file.Reader(nil, trx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "can't read a directory")
}

func TestFile_Reader2(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	tempDir := filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	randomBytes := Random(ChunkSize*2 + 3)
	randomBytesReader := bytes.NewReader(randomBytes)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	file, err := CreateFileFromReader(app, "/test/random.bytes", randomBytesReader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, file.Object.Hash)

	reader, err := file.Reader(&tempDir, trx)
	assert.Nil(t, err)
	allContent, err := ioutil.ReadAll(reader)
	assert.Nil(t, err)
	allContentHash, err := util.Sha256Hash2String(allContent)
	assert.Nil(t, err)
	assert.Equal(t, allContentHash, file.Object.Hash)
}

func TestFindFileByUID(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	file, err := CreateOrGetLastDirectory(app, "/test", trx)
	assert.Nil(t, err)

	file1, err := FindFileByUID(file.UID, false, trx)
	assert.Nil(t, err)
	assert.Equal(t, file1.ID, file.ID)

	assert.Nil(t, trx.Delete(file).Error)

	_, err = FindFileByUID(file.UID, false, trx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "record not found")

	file2, err := FindFileByUID(file.UID, true, trx)
	assert.Nil(t, err)
	assert.Equal(t, file2.ID, file.ID)
}

func TestFile_CanBeAccessedByToken(t *testing.T) {
	token, trx, down, err := NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	dir, err := CreateOrGetLastDirectory(&token.App, "/test/create/a/directory", trx)
	assert.Nil(t, err)

	token.Path = "/test"
	assert.Nil(t, trx.Save(token).Error)
	assert.Nil(t, dir.CanBeAccessedByToken(token, trx))

	token.Path = "/create"
	assert.Nil(t, trx.Save(token).Error)
	assert.Equal(t, dir.CanBeAccessedByToken(token, trx), ErrAccessDenied)
}

// TestFile_MoveTo is used to test move file
func TestFile_MoveTo(t *testing.T) {
	var (
		err               error
		app               *App
		trx               *gorm.DB
		file              *File
		down              func(*testing.T)
		tempDir           = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		rootDir           *File
		randomBytes       = Random(255)
		randomBytesReader = bytes.NewReader(randomBytes)
	)
	app, trx, down, err = newAppForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	file, err = CreateFileFromReader(app, "/save/to/a/1.bytes", randomBytesReader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 255, file.Size)
	assert.Equal(t, 255, file.Parent.Size)
	assert.Equal(t, "/save/to/a/1.bytes", file.mustPath(trx))
	aDir := file.Parent
	assert.Equal(t, 255, aDir.Size)
	rootDir, _ = CreateOrGetRootPath(app, trx)
	assert.Equal(t, 255, rootDir.Size)

	// only rename
	err = file.MoveTo("/save/to/a/2.bytes", trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/a/2.bytes", file.mustPath(trx))
	assert.Equal(t, aDir.ID, file.Parent.ID)
	rootDir, err = CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 255, rootDir.Size)

	// move to another dir
	err = file.MoveTo("/save/to/b/2.bytes", trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/b/2.bytes", file.mustPath(trx))
	assert.NotEqual(t, aDir.ID, file.Parent.ID)
	bDir := file.Parent
	assert.Equal(t, file.Parent.ID, bDir.ID)
	assert.Equal(t, 255, bDir.Size)

	// nothing change
	err = file.MoveTo("/save/to/b/2.bytes", trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/b/2.bytes", file.mustPath(trx))

	rootDir, err = CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 255, rootDir.Size)
}

// TestFile_MoveTo2 is used to move a directory
func TestFile_MoveTo2(t *testing.T) {
	var (
		err               error
		app               *App
		trx               *gorm.DB
		file              *File
		down              func(*testing.T)
		tempDir           = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		rootDir           *File
		randomBytes       = Random(255)
		randomBytesReader = bytes.NewReader(randomBytes)
	)
	app, trx, down, err = newAppForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	file, err = CreateFileFromReader(app, "/save/to/a/1.bytes", randomBytesReader, int8(0), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/to/a/1.bytes", file.mustPath(trx))
	dir := file.Parent

	// rename directory
	err = dir.MoveTo("/save/to/b", trx)
	assert.Nil(t, err)
	assert.Equal(t, "b", dir.Name)
	assert.Equal(t, int8(1), dir.IsDir)
	assert.Equal(t, "/save/to/b/1.bytes", file.mustPath(trx))

	// move to another directory
	err = dir.MoveTo("/save/as/b", trx)
	assert.Nil(t, err)
	assert.Equal(t, "/save/as/b/1.bytes", file.mustPath(trx))

	rootDir, err = CreateOrGetRootPath(app, trx)
	assert.Nil(t, err)
	assert.Equal(t, 255, rootDir.Size)

	saveToDir, err := FindFileByPath(app, "/save/to", trx)
	assert.Nil(t, err)
	assert.Equal(t, 0, saveToDir.Size)

	saveAsDir, err := FindFileByPath(app, "/save/as", trx)
	assert.Nil(t, err)
	assert.Equal(t, 255, saveAsDir.Size)
}
