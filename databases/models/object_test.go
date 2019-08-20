//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"labix.org/v2/mgo/bson"

	"github.com/bigfile/bigfile/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestObject_TableName(t *testing.T) {
	assert.Equal(t, Object{}.TableName(), "objects")
}

func TestObject_ChunkCount(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)
	assert.Equal(t, 0, object.ChunkCount(trx))
}

func TestObject_ChunkCount2(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
		ObjectChunks: []ObjectChunk{
			{
				Chunk: Chunk{
					Size: size,
					Hash: hash,
				},
				Number: 1,
			},
		},
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)
	assert.Equal(t, 1, object.ChunkCount(trx))
}

func TestObject_LastChunk(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)
	assert.Equal(t, 0, object.ChunkCount(trx))
	_, err = object.LastChunk(trx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestObject_LastChunk2(t *testing.T) {
	var (
		content  = "hello world"
		content2 = "make money"
		size     = len(content)
		size2    = len(content2)
		err      error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	hash2, err := util.Sha256Hash2String([]byte(content2))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
		ObjectChunks: []ObjectChunk{
			{
				Chunk: Chunk{
					Size: size,
					Hash: hash,
				},
				Number: 1,
			},
			{
				Chunk: Chunk{
					Size: size2,
					Hash: hash2,
				},
				Number: 2,
			},
		},
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)
	assert.Equal(t, 2, object.ChunkCount(trx))

	chunk, err := object.LastChunk(trx)
	assert.Nil(t, err)
	assert.Equal(t, chunk.Size, size2)
}

func TestObject_LastChunkNumber(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
		number  int
		object  *Object
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object = &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)
	assert.Equal(t, 0, object.ChunkCount(trx))
	number, err = object.LastChunkNumber(trx)
	assert.Nil(t, err)
	assert.Equal(t, 0, number)
}

func TestObject_LastChunkNumber2(t *testing.T) {
	var (
		content  = "hello world"
		content2 = "make money"
		size     = len(content)
		size2    = len(content2)
		err      error
		number   int
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	hash2, err := util.Sha256Hash2String([]byte(content2))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
		ObjectChunks: []ObjectChunk{
			{
				Chunk: Chunk{
					Size: size,
					Hash: hash,
				},
				Number: 1,
			},
			{
				Chunk: Chunk{
					Size: size2,
					Hash: hash2,
				},
				Number: 2,
			},
		},
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)
	assert.Equal(t, 2, object.ChunkCount(trx))

	number, err = object.LastChunkNumber(trx)
	assert.Nil(t, err)
	assert.Equal(t, 2, number)
}

func TestObject_LastObjectChunk(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)

	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)

	oc1 := &ObjectChunk{
		ObjectID: object.ID,
		ChunkID:  1,
		Number:   1,
	}
	assert.Nil(t, trx.Create(oc1).Error)

	oc2 := &ObjectChunk{
		ObjectID: object.ID,
		ChunkID:  1,
		Number:   2,
	}
	assert.Nil(t, trx.Create(oc2).Error)

	ocTmp, err := object.LastObjectChunk(trx)
	assert.Nil(t, err)
	assert.Equal(t, ocTmp.ID, oc2.ID)
}

func TestFindObjectByHash(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	objectTmp, err := FindObjectByHash(hash, trx)
	assert.Nil(t, err)
	assert.Equal(t, objectTmp.ID, object.ID)
}

func TestCreateObjectFromReader(t *testing.T) {
	var (
		tempDir   = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		randomStr = Random(uint(ChunkSize * 2.5))
		reader    = strings.NewReader(string(randomStr))
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	object, err := CreateObjectFromReader(reader, &tempDir, trx)
	assert.Nil(t, err)
	defer func() {
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
		down(t)
	}()
	hash, err := util.Sha256Hash2String(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, hash, object.Hash)
	assert.Equal(t, int(ChunkSize*2.5), object.Size)
}

func TestObject_FileCount(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	file1 := &File{
		UID:      bson.NewObjectId().Hex(),
		ObjectID: object.ID,
	}
	assert.Nil(t, trx.Save(file1).Error)
	assert.Equal(t, object.FileCount(trx), 1)
}

func TestObject_FileCount2(t *testing.T) {
	var (
		content = "hello world"
		size    = len(content)
		err     error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	hash, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: hash,
	}
	assert.Nil(t, trx.Save(object).Error)
	assert.Equal(t, object.FileCount(trx), 0)
}
