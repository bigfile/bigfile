//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package util

import (
	"errors"
	"os"
	"reflect"
)

// IsDir is used to judge whether the specific path is a valid directory
func IsDir(path string) bool {
	var (
		fileInfo os.FileInfo
		err      error
	)
	fileInfo, err = os.Stat(path)

	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// IsFile is used to judge whether the specific path is a valid file path
func IsFile(path string) bool {
	var (
		fileInfo os.FileInfo
		err      error
	)
	fileInfo, err = os.Stat(path)

	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

// SubStrFromTo will truncate the specific string by parameters
func SubStrFromTo(s string, from, to int) string {
	if from < 0 {
		from = len(s) + from
	}
	if to < 0 {
		to = len(s) + to
	}
	return s[from:to]
}

// SubStrFromToEnd will truncate the specific string by parameters
func SubStrFromToEnd(s string, from int) string {
	return SubStrFromTo(s, from, len(s))
}

// ReverseSlice is used to reverse a slice
func ReverseSlice(data interface{}) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Slice {
		panic(errors.New("data must be a slice type"))
	}
	valueLen := value.Len()
	swap := reflect.Swapper(data)
	for i := 0; i <= int((valueLen-1)/2); i++ {
		reverseIndex := valueLen - 1 - i
		swap(i, reverseIndex)
	}
}
