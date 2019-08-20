//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestChunk_TableName(t *testing.T) {
	assert.Equal(t, (Chunk{}).TableName(), "chunks")
}

func TestChunk_Path(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
		assert.Contains(t, err.(error).Error(), "invalid chunk id")
	}()
	chunk := &Chunk{}
	chunk.Path(nil)
}

func TestChunk_Path2(t *testing.T) {

	tempDir := filepath.Join(os.TempDir(), RandomWithMd5(32))
	defer func() {
		os.RemoveAll(tempDir)
	}()

	chunk := &Chunk{ID: 100044}
	assert.Equal(t, strings.TrimPrefix(chunk.Path(&tempDir), tempDir), "/100/100044")
	chunk = &Chunk{ID: 10001}
	assert.Equal(t, strings.TrimPrefix(chunk.Path(&tempDir), tempDir), "/10/10001")
	chunk = &Chunk{ID: 100001}
	assert.Equal(t, strings.TrimPrefix(chunk.Path(&tempDir), tempDir), "/100/100001")
}

func TestFindChunkByHash(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)

	var (
		str     = []byte(RandomWithMd5(256))
		strLen  = len(str)
		err     error
		strHash string
		chunk   *Chunk
	)

	strHash, err = util.Sha256Hash2String(str)
	assert.Nil(t, err)
	chunk = &Chunk{Size: strLen, Hash: strHash}
	err = trx.Create(chunk).Error
	assert.Nil(t, err)

	chunkTmp, err := FindChunkByHash(strHash, trx)
	assert.Nil(t, err)
	assert.Equal(t, chunk.ID, chunkTmp.ID)
}

func TestCreateChunkFromBytes(t *testing.T) {
	var (
		bigBytes = []byte(strings.Repeat("s", ChunkSize+1))
		tempDir  = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		trx      *gorm.DB
		err      error
		down     func(*testing.T)
		chunk    *Chunk
	)

	trx, down = setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	_, err = CreateChunkFromBytes(bigBytes, &tempDir, trx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "the size of chunk must be less than")

	bigBytes = bigBytes[:ChunkSize]
	chunk, err = CreateChunkFromBytes(bigBytes, &tempDir, trx)
	assert.Nil(t, err)
	assert.True(t, chunk.ID > 0)

	path := chunk.Path(&tempDir)
	assert.True(t, util.IsFile(path))

	fileInfo, err := os.Stat(path)
	assert.Nil(t, err)
	assert.Equal(t, fileInfo.Size(), int64(ChunkSize))
}

func TestChunk_AppendBytes(t *testing.T) {
	var (
		bigBytes   = []byte("hello")
		tempDir    = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		trx        *gorm.DB
		err        error
		down       func(*testing.T)
		chunk      *Chunk
		writeCount int
	)

	trx, down = setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	chunk, err = CreateChunkFromBytes(bigBytes, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 5, chunk.Size)
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", chunk.Hash)

	_, writeCount, err = chunk.AppendBytes([]byte(" world"), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 6, writeCount)
	assert.Equal(t, 11, chunk.Size)
	assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", chunk.Hash)
}

func TestChunk_AppendBytes2(t *testing.T) {
	var (
		bigBytes     = []byte("hello")
		tempDir      = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		trx          *gorm.DB
		err          error
		down         func(*testing.T)
		chunk        *Chunk
		writeCount   int
		contentBytes = []byte("hello world")
	)

	trx, down = setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	chunk, err = CreateChunkFromBytes(bigBytes, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 5, chunk.Size)
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", chunk.Hash)

	chunk2, err := CreateChunkFromBytes(contentBytes, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 11, chunk2.Size)
	assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", chunk2.Hash)

	chunkTmp, writeCount, err := chunk.AppendBytes([]byte(" world"), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 6, writeCount)
	assert.Equal(t, 11, chunkTmp.Size)
	assert.Equal(t, chunkTmp.ID, chunk2.ID)
}

func TestChunk_AppendBytes3(t *testing.T) {
	var (
		bigBytes   = []byte("hello")
		tempDir    = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
		trx        *gorm.DB
		err        error
		down       func(*testing.T)
		chunk      *Chunk
		writeCount int
	)

	trx, down = setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	chunk, err = CreateChunkFromBytes(bigBytes, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 5, chunk.Size)
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", chunk.Hash)

	assert.Nil(t, trx.Create(&ObjectChunk{
		ObjectID: 1,
		ChunkID:  chunk.ID,
		Number:   1,
	}).Error)

	assert.Nil(t, trx.Create(&ObjectChunk{
		ObjectID: 2,
		ChunkID:  chunk.ID,
		Number:   1,
	}).Error)

	newChunk, writeCount, err := chunk.AppendBytes([]byte(" world"), &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, 6, writeCount)
	assert.Equal(t, 11, newChunk.Size)
	assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", newChunk.Hash)
	assert.True(t, newChunk.ID != chunk.ID)
}

func TestCreateEmptyContentChunk(t *testing.T) {
	var (
		trx     *gorm.DB
		err     error
		down    func(*testing.T)
		chunk   *Chunk
		tempDir = filepath.Join(os.TempDir(), strconv.FormatInt(rand.Int63n(1<<32), 10))
	)
	trx, down = setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	chunk, err = CreateEmptyContentChunk(&tempDir, trx)
	assert.Nil(t, err)
	assert.True(t, chunk.ID > 0)
	fmt.Println(tempDir)
}
