//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
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

func TestSubStrFromTo(t *testing.T) {
	s := "hello world"
	assert.Equal(t, s[1:3], SubStrFromTo(s, 1, 3))
	assert.Equal(t, s[0:11], SubStrFromTo(s, 0, 11))
	assert.Equal(t, s[8:10], SubStrFromTo(s, -3, -1))
	assert.Equal(t, s[:8], SubStrFromTo(s, 0, -3))
}

func TestSubStrFromToEnd(t *testing.T) {
	s := "hello world"
	assert.Equal(t, s[1:], SubStrFromToEnd(s, 1))
	assert.Equal(t, s[5:], SubStrFromToEnd(s, -6))
}

func TestSubStrFromToEnd2(t *testing.T) {
	s := "hello world"
	assert.Equal(t, s, SubStrFromTo(s, 0, -6)+SubStrFromToEnd(s, -6))
}

func TestReverseSlice(t *testing.T) {

	names := []string{"bob", "mary"}
	ReverseSlice(names)
	reflect.DeepEqual(names, []string{"mary", "bob"})

	names = []string{"bob", "mary", "michael"}
	ReverseSlice(names)
	reflect.DeepEqual(names, []string{"michael", "mary", "bob"})

	ages := []int{24, 26}
	ReverseSlice(ages)
	reflect.DeepEqual(ages, []int{26, 24})

	ages = []int{24, 26, 28}
	ReverseSlice(ages)
	reflect.DeepEqual(ages, []int{28, 26, 24})
}
