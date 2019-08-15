//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsDir(t *testing.T) {
	assert.True(t, IsDir(os.TempDir()))
	assert.False(t, IsDir(fmt.Sprintf("%s/%d", os.TempDir(), rand.Int63n(100000000000))))
}

func TestIsFile(t *testing.T) {
	filepath.Join(os.TempDir(), "")
	file, err := afero.TempFile(afero.NewOsFs(), os.TempDir(), fmt.Sprintf("%d", rand.Int63n(1<<32)))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	assert.True(t, IsFile(file.Name()))
	assert.False(t, IsFile(filepath.Join(os.TempDir(), fmt.Sprintf("%d", rand.Int63n(1<<32)))))
}
