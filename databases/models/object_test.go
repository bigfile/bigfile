//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"bytes"
	sha2562 "crypto/sha256"
	"encoding/hex"
	"hash"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/internal/sha256"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/stretchr/testify/assert"
	"labix.org/v2/mgo/bson"
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{Size: size, Hash: h}
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: h,
		ObjectChunks: []ObjectChunk{
			{
				Chunk: Chunk{
					Size: size,
					Hash: h,
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{Size: size, Hash: h}
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	hash2, err := util.Sha256Hash2String([]byte(content2))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: h,
		ObjectChunks: []ObjectChunk{
			{
				Chunk: Chunk{
					Size: size,
					Hash: h,
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object = &Object{Size: size, Hash: h}
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	hash2, err := util.Sha256Hash2String([]byte(content2))
	assert.Nil(t, err)
	object := &Object{
		Size: size,
		Hash: h,
		ObjectChunks: []ObjectChunk{
			{
				Chunk: Chunk{
					Size: size,
					Hash: h,
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{Size: size, Hash: h}
	assert.Nil(t, trx.Save(object).Error)
	assert.True(t, object.ID > 0)

	oc1 := &ObjectChunk{ObjectID: object.ID, ChunkID: 1, Number: 1}
	assert.Nil(t, trx.Create(oc1).Error)

	oc2 := &ObjectChunk{ObjectID: object.ID, ChunkID: 1, Number: 2}
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{Size: size, Hash: h}
	assert.Nil(t, trx.Save(object).Error)
	objectTmp, err := FindObjectByHash(h, trx)
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
	h, err := util.Sha256Hash2String(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, h, object.Hash)
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{Size: size, Hash: h}
	assert.Nil(t, trx.Save(object).Error)
	file1 := &File{UID: bson.NewObjectId().Hex(), ObjectID: object.ID}
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
	h, err := util.Sha256Hash2String([]byte(content))
	assert.Nil(t, err)
	object := &Object{Size: size, Hash: h}
	assert.Nil(t, trx.Save(object).Error)
	assert.Equal(t, object.FileCount(trx), 0)
}

func TestCreateEmptyObject(t *testing.T) {
	var (
		oc               *ObjectChunk
		err              error
		chunk            *Chunk
		object           *Object
		tempDir          = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		stateHash        hash.Hash
		emptyContentHash string
	)
	emptyContentHash, err = util.Sha256Hash2String(nil)
	assert.Nil(t, err)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	object, err = CreateEmptyObject(&tempDir, trx)
	assert.Nil(t, err)
	assert.True(t, object.ID > 0)
	assert.Equal(t, emptyContentHash, object.Hash)
	chunk, err = object.LastChunk(trx)
	assert.Nil(t, err)
	assert.Equal(t, emptyContentHash, chunk.Hash)
	oc, err = object.LastObjectChunk(trx)
	assert.Nil(t, err)
	stateHash, err = sha256.NewHashWithStateText(*oc.HashState)
	assert.Nil(t, err)
	assert.Equal(t, emptyContentHash, hex.EncodeToString(stateHash.Sum(nil)))
}

// TestObject_AppendFromReader is used to test append content to an object
func TestObject_AppendFromReader(t *testing.T) {
	var (
		h         = sha2562.New()
		oc        *ObjectChunk
		err       error
		size      int
		tempDir   = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		randomStr = Random(uint(ChunkSize * 2.5))
		stateHash hash.Hash
		object    *Object
		reader    = strings.NewReader(string(randomStr))
		prevSize  int
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	object, err = CreateObjectFromReader(reader, &tempDir, trx)
	assert.Nil(t, err)
	defer func() {
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
		down(t)
	}()
	_, err = h.Write(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), object.Hash)
	assert.Equal(t, int(ChunkSize*2.5), object.Size)
	assert.Equal(t, 3, object.ChunkCount(trx))

	randomStr = Random(uint(ChunkSize * 0.5))
	prevSize = object.Size
	object, size, err = object.AppendFromReader(bytes.NewReader(randomStr), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, int(ChunkSize*0.5), size)
	assert.Equal(t, prevSize+int(ChunkSize*0.5), object.Size)
	_, err = h.Write(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), object.Hash)
	assert.Equal(t, 3, object.ChunkCount(trx))
	oc, err = object.LastObjectChunk(trx)
	assert.Nil(t, err)
	stateHash, err = sha256.NewHashWithStateText(*oc.HashState)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, hex.EncodeToString(stateHash.Sum(nil)))

	chunkSize := ChunkSize
	randomStr = Random(uint(float64(chunkSize) * 0.12))
	prevSize = object.Size
	object, size, err = object.AppendFromReader(bytes.NewReader(randomStr), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, int(float64(chunkSize)*0.12), size)
	assert.Equal(t, int(float64(chunkSize)*0.12)+prevSize, object.Size)
	_, err = h.Write(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), object.Hash)
	assert.Equal(t, 4, object.ChunkCount(trx))
	oc, err = object.LastObjectChunk(trx)
	assert.Nil(t, err)
	assert.Equal(t, 4, oc.Number)
	stateHash, err = sha256.NewHashWithStateText(*oc.HashState)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, hex.EncodeToString(stateHash.Sum(nil)))
}

// TestObject_AppendFromReader2 is used to test that an object will be appended
// be referenced by many files
func TestObject_AppendFromReader2(t *testing.T) {
	var (
		h                 = sha2562.New()
		oc                *ObjectChunk
		err               error
		size              int
		tempDir           = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		randomStr         = Random(uint(ChunkSize * 2.5))
		stateHash         hash.Hash
		object            *Object
		object2           *Object
		reader            = strings.NewReader(string(randomStr))
		originContentHash string
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	object, err = CreateObjectFromReader(reader, &tempDir, trx)
	assert.Nil(t, err)
	defer func() {
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
		down(t)
	}()
	_, err = h.Write(randomStr)
	assert.Nil(t, err)
	originContentHash = hex.EncodeToString(h.Sum(nil))
	assert.Equal(t, originContentHash, object.Hash)
	assert.Equal(t, int(ChunkSize*2.5), object.Size)
	assert.Equal(t, 3, object.ChunkCount(trx))
	oc, err = object.LastObjectChunk(trx)
	assert.Nil(t, err)
	stateHash, err = sha256.NewHashWithStateText(*oc.HashState)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, hex.EncodeToString(stateHash.Sum(nil)))

	file1 := &File{UID: bson.NewObjectId().Hex(), ObjectID: object.ID, Name: "file1"}
	file2 := &File{UID: bson.NewObjectId().Hex(), ObjectID: object.ID, Name: "file2"}
	assert.Nil(t, trx.Save(file1).Error)
	assert.Nil(t, trx.Save(file2).Error)

	chunkSize := ChunkSize
	randomStr = Random(uint(float64(chunkSize) * 0.12))
	object2, size, err = object.AppendFromReader(bytes.NewReader(randomStr), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, int(float64(chunkSize)*0.12), size)
	_, err = h.Write(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), object2.Hash)
	assert.Equal(t, 3, object2.ChunkCount(trx))
	oc, err = object2.LastObjectChunk(trx)
	assert.Nil(t, err)
	stateHash, err = sha256.NewHashWithStateText(*oc.HashState)
	assert.Nil(t, err)
	assert.Equal(t, object2.Hash, hex.EncodeToString(stateHash.Sum(nil)))

	assert.Equal(t, 3, object.ChunkCount(trx))
	assert.Equal(t, originContentHash, object.Hash)
	assert.NotEqual(t, object.ID, object2.ID)
}

// TestObject_AppendFromReader3 is used to test an object that consists of multiple
// the same chunk
func TestObject_AppendFromReader3(t *testing.T) {
	var (
		h         = sha2562.New()
		oc        *ObjectChunk
		err       error
		size      int
		tempDir   = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		randomStr = Random(uint(ChunkSize))
		stateHash hash.Hash
		object    *Object
		object2   *Object
		reader    = bytes.NewBuffer(randomStr)
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	object, err = CreateObjectFromReader(reader, &tempDir, trx)
	assert.Nil(t, err)
	defer func() {
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
		down(t)
	}()
	_, err = h.Write(randomStr)
	assert.Nil(t, err)
	assert.Equal(t, hex.EncodeToString(h.Sum(nil)), object.Hash)
	assert.Equal(t, ChunkSize, object.Size)
	assert.Equal(t, 1, object.ChunkCount(trx))
	oc, err = object.LastObjectChunk(trx)
	assert.Nil(t, err)
	stateHash, err = sha256.NewHashWithStateText(*oc.HashState)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, hex.EncodeToString(stateHash.Sum(nil)))

	object2, size, err = object.AppendFromReader(bytes.NewReader(randomStr), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, ChunkSize, size)
	assert.Equal(t, object2.ID, object.ID)
	assert.Equal(t, 2, object2.ChunkCount(trx))
	oc2, err := object2.LastObjectChunk(trx)
	assert.Nil(t, err)
	assert.Equal(t, oc.ChunkID, oc2.ChunkID)
}

func TestObject_Reader(t *testing.T) {
	object, rootPath, down := newObjectForObjectReaderTest(t)
	defer down(t)
	_, err := object.Reader(rootPath)
	assert.Nil(t, err)
}
