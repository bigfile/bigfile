//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package ftp

import (
	"os"
	"time"
)

// FileInfo is used to represent the information of file or directory
type FileInfo struct {
	name     string
	size     int64
	isDir    bool
	modeTime time.Time
}

func (f *FileInfo) Sys() interface{} {
	return nil
}

// IsDir represent whether the object is a directory
func (f *FileInfo) IsDir() bool {
	return f.isDir
}

// ModTime is used to return the modify time of file
func (f *FileInfo) ModTime() time.Time {
	return f.modeTime
}

// Mode returns a file's mode and permission bits.
func (f *FileInfo) Mode() os.FileMode {
	if f.isDir {
		return os.ModePerm | os.ModeDir
	}
	return os.ModePerm
}

// Size return the size of file or directory
func (f *FileInfo) Size() int64 {
	return f.size
}

// Name return the name of file or directory
func (f *FileInfo) Name() string {
	return f.name
}

// Owner return the owner of file
func (f *FileInfo) Owner() string {
	return "bigfile"
}

// Owner return the group of file
func (f *FileInfo) Group() string {
	return "bigfile"
}
