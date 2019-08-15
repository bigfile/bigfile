//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"testing"

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
	chunk.Path()
}

func TestChunk_Path2(t *testing.T) {
	chunk := &Chunk{ID: 10000223344}
	fmt.Println(chunk.Path())
}
