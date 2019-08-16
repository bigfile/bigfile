//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
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
