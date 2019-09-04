//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package ftp

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileInfo_Group(t *testing.T) {
	assert.Equal(t, "bigfile", (&FileInfo{}).Group())
}

func TestFileInfo_IsDir(t *testing.T) {
	assert.True(t, (&FileInfo{isDir: true}).IsDir())
}

func TestFileInfo_Mode(t *testing.T) {
	assert.Equal(t, os.ModePerm|os.ModeDir, (&FileInfo{isDir: true}).Mode())
	assert.Equal(t, os.ModePerm, (&FileInfo{isDir: false}).Mode())
}

func TestFileInfo_ModTime(t *testing.T) {
	modTime := time.Now()
	assert.Equal(t, modTime, (&FileInfo{modeTime: modTime}).ModTime())
}

func TestFileInfo_Name(t *testing.T) {
	assert.Equal(t, "test", (&FileInfo{name: "test"}).Name())
}

func TestFileInfo_Owner(t *testing.T) {
	assert.Equal(t, "bigfile", (&FileInfo{}).Owner())
}

func TestFileInfo_Size(t *testing.T) {
	assert.Equal(t, int64(222), (&FileInfo{size: 222}).Size())
}

func TestFileInfo_Sys(t *testing.T) {
	assert.Nil(t, (&FileInfo{}).Sys())
}
