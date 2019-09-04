//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectReader(t *testing.T) {

	_, err := NewObjectReader(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, ErrInvalidObject)

	db, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)

	object := &Object{}
	assert.Nil(t, db.Save(object).Error)

	_, err = NewObjectReader(object, nil, db)
	assert.NotNil(t, err)
	assert.Equal(t, err, ErrObjectNoChunks)
}

func newObjectForObjectReaderTest(t *testing.T) (*Object, *string, func(*testing.T), *gorm.DB) {
	var (
		err               error
		object            *Object
		tempDir           = NewTempDirForTest()
		randomBytes       = Random(ChunkSize*2 + 123)
		randomBytesHash   string
		randomBytesReader = bytes.NewReader(randomBytes)
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	object, err = CreateObjectFromReader(randomBytesReader, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, randomBytesHash)
	assert.Nil(t, trx.Preload("Chunks").Find(object).Error)
	return object, &tempDir, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}, trx
}

func TestNewObjectReader2(t *testing.T) {
	object, rootPath, down, trx := newObjectForObjectReaderTest(t)
	defer down(t)
	_, err := NewObjectReader(object, rootPath, trx)
	assert.Nil(t, err)
}

func TestObjectReader_Read(t *testing.T) {
	object, rootPath, down, trx := newObjectForObjectReaderTest(t)
	defer down(t)
	or, err := NewObjectReader(object, rootPath, trx)
	assert.Nil(t, err)

	allContent, err := ioutil.ReadAll(or)
	assert.Nil(t, err)
	allContentHash, err := util.Sha256Hash2String(allContent)
	assert.Nil(t, err)
	assert.Equal(t, allContentHash, object.Hash)
}

func TestObjectReader_Seek(t *testing.T) {
	var (
		err               error
		object            *Object
		tempDir           = NewTempDirForTest()
		randomBytes       = []byte("this is a random string, only used here")
		randomBytesHash   string
		randomBytesReader = bytes.NewReader(randomBytes)
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	object, err = CreateObjectFromReader(randomBytesReader, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, randomBytesHash)

	// seek relative to start
	bytesContent := randomBytes[10:16]
	or, err := NewObjectReader(object, &tempDir, trx)
	assert.Nil(t, err)
	offset, err := or.Seek(10, io.SeekStart)
	assert.Nil(t, err)
	assert.Equal(t, int64(10), offset)
	readContents := make([]byte, 6)
	readCount, err := or.Read(readContents)
	assert.Nil(t, err)
	assert.Equal(t, readCount, 6)
	assert.True(t, bytes.Equal(bytesContent, readContents))

	// seek relative to current
	offset, err = or.Seek(7, io.SeekCurrent)
	assert.Nil(t, err)
	assert.Equal(t, 23, int(offset))

	// seek relative to end
	offset, err = or.Seek(10, io.SeekEnd)
	assert.Nil(t, err)
	assert.Equal(t, offset, int64(len(randomBytes)+10))
	_, err = or.Read(nil)
	assert.Equal(t, io.EOF, err)

	// invalid whence
	_, err = or.Seek(10, -1)
	assert.Equal(t, ErrInvalidSeekWhence, err)

	// negative seek
	_, err = or.Seek(-1, io.SeekStart)
	assert.Equal(t, ErrNegativePosition, err)
}

// the offset is integer multiple to ChunkSize
func TestObjectReader_Seek2(t *testing.T) {
	var (
		err               error
		object            *Object
		tempDir           = NewTempDirForTest()
		randomBytes       = Random(ChunkSize + 1024)
		randomBytesHash   string
		randomBytesReader = bytes.NewReader(randomBytes)
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	object, err = CreateObjectFromReader(randomBytesReader, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, randomBytesHash)

	bytesContent := randomBytes[ChunkSize:]
	bytesContentHash, err := util.Sha256Hash2String(bytesContent)
	assert.Nil(t, err)
	or, err := NewObjectReader(object, &tempDir, trx)
	assert.Nil(t, err)
	offset, err := or.Seek(int64(ChunkSize), io.SeekStart)
	assert.Nil(t, err)
	assert.Equal(t, int64(ChunkSize), offset)
	restContent, err := ioutil.ReadAll(or)
	assert.Nil(t, err)
	restContentHash, err := util.Sha256Hash2String(restContent)
	assert.Nil(t, err)
	assert.Equal(t, bytesContentHash, restContentHash)

	bytesContent = randomBytes[ChunkSize-1:]
	bytesContentHash, err = util.Sha256Hash2String(bytesContent)
	assert.Nil(t, err)
	or, err = NewObjectReader(object, &tempDir, trx)
	assert.Nil(t, err)
	offset, err = or.Seek(int64(ChunkSize-1), io.SeekStart)
	assert.Nil(t, err)
	assert.Equal(t, int64(ChunkSize-1), offset)
	restContent, err = ioutil.ReadAll(or)
	assert.Nil(t, err)
	restContentHash, err = util.Sha256Hash2String(restContent)
	assert.Nil(t, err)
	assert.Equal(t, bytesContentHash, restContentHash)
}
