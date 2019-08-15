//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"os"
	"path/filepath"
	"strings"
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
