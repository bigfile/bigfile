//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectChunk_TableName(t *testing.T) {
	assert.Equal(t, "object_chunk", ObjectChunk{}.TableName())
}

func TestCountObjectChunkByChunkId(t *testing.T) {

	var (
		count int
		err   error
	)

	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)

	assert.Nil(t, trx.Create(&ObjectChunk{
		ObjectID: 1,
		ChunkID:  1,
		Number:   1,
	}).Error)

	assert.Nil(t, trx.Create(&ObjectChunk{
		ObjectID: 2,
		ChunkID:  1,
		Number:   1,
	}).Error)

	assert.Nil(t, trx.Create(&ObjectChunk{
		ObjectID: 3,
		ChunkID:  2,
		Number:   1,
	}).Error)

	count, err = CountObjectChunkByChunkID(1, trx)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	count, err = CountObjectChunkByChunkID(2, trx)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	count, err = CountObjectChunkByChunkID(3, trx)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}
