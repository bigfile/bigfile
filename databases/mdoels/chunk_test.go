//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestChunk_TableName(t *testing.T) {
	assert.Equal(t, (Chunk{}).TableName(), "chunks")
}
