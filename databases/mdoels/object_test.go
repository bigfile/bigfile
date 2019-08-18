//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"testing"

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
