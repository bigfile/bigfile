//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestNewSecret(t *testing.T) {
	assert.Equal(t, len(NewSecret()), 32)
}

func BenchmarkNewSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSecret()
	}
}

func TestRandom(t *testing.T) {
	assert.Equal(t, 128, len(Random(128)))
}

func BenchmarkRandom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Random(32)
	}
}

func TestRandomWithMd5(t *testing.T) {
	assert.Equal(t, len(RandomWithMd5(128)), 32)
}
