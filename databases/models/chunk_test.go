//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"io/ioutil"
	"os"
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
	chunk := &Chunk{}
	_, err := chunk.Path(nil)
	assert.Equal(t, ErrInvalidChunkID, err)
}

func TestChunk_Path2(t *testing.T) {
	var (
		err     error
		path    string
		tempDir = NewTempDirForTest()
	)
	defer func() { os.RemoveAll(tempDir) }()

	chunk := &Chunk{ID: 100044}
	path, err = chunk.Path(&tempDir)
	assert.Nil(t, err)
	assert.Equal(t, strings.TrimPrefix(path, tempDir), "/100/100044")
	chunk = &Chunk{ID: 10001}
	path, err = chunk.Path(&tempDir)
	assert.Nil(t, err)
	assert.Equal(t, strings.TrimPrefix(path, tempDir), "/10/10001")
	chunk = &Chunk{ID: 100001}
	path, err = chunk.Path(&tempDir)
	assert.Nil(t, err)
	assert.Equal(t, strings.TrimPrefix(path, tempDir), "/100/100001")
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
		tempDir  = NewTempDirForTest()
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

	path, err := chunk.Path(&tempDir)
	assert.Nil(t, err)
	assert.True(t, util.IsFile(path))

	fileInfo, err := os.Stat(path)
	assert.Nil(t, err)
	assert.Equal(t, fileInfo.Size(), int64(ChunkSize))
}

func TestChunk_AppendBytes(t *testing.T) {
	var (
		bigBytes   = []byte("hello")
		tempDir    = NewTempDirForTest()
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

	_, _, err = chunk.AppendBytes(Random(ChunkSize), &tempDir, trx)
	assert.NotNil(t, err)
	assert.Equal(t, err, ErrChunkExceedLimit)
}

func TestChunk_AppendBytes2(t *testing.T) {
	var (
		bigBytes     = []byte("hello")
		tempDir      = NewTempDirForTest()
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
		tempDir    = NewTempDirForTest()
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
		tempDir = NewTempDirForTest()
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

func TestChunk_Reader(t *testing.T) {
	var (
		trx         *gorm.DB
		err         error
		down        func(*testing.T)
		chunk       *Chunk
		tempDir     = NewTempDirForTest()
		tempDir2    = NewTempDirForTest()
		randomBytes = Random(ChunkSize)
	)

	trx, down = setUpTestCaseWithTrx(nil, t)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
		if util.IsDir(tempDir2) {
			os.RemoveAll(tempDir2)
		}
	}()

	chunk, err = CreateChunkFromBytes(randomBytes, &tempDir, trx)
	assert.Nil(t, err)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, chunk.Hash)

	chunkReader, err := chunk.Reader(&tempDir)
	assert.Nil(t, err)
	allContent, err := ioutil.ReadAll(chunkReader)
	assert.Nil(t, err)
	allContentHash, err := util.Sha256Hash2String(allContent)
	assert.Nil(t, err)
	assert.Equal(t, allContentHash, randomBytesHash)

	_, err = chunk.Reader(&tempDir2)
	assert.NotNil(t, err)
}
