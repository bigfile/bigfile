//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package util

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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

func TestSha256Hash2String(t *testing.T) {
	s := "hello world"
	h := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	sh, err := Sha256Hash2String([]byte(s))
	assert.Nil(t, err)
	assert.Equal(t, sh, h)
}

func TestSha256Hash2String2(t *testing.T) {
	sh, err := Sha256Hash2String(nil)
	assert.Nil(t, err)
	h := sha256.New()
	assert.Equal(t, sh, hex.EncodeToString(h.Sum(nil)))
}

func TestIsRecordNotFound(t *testing.T) {
	assert.False(t, IsRecordNotFound(errors.New("")))
	assert.True(t, IsRecordNotFound(errors.New("record not found")))
}

func TestInTransaction(t *testing.T) {
	dbFile := filepath.Join(os.TempDir(), fmt.Sprintf("%d-gorm.db", rand.Int63()))
	db, err := gorm.Open("sqlite3", dbFile)
	assert.Nil(t, err)
	defer func() {
		_ = db.Close()
		if IsFile(dbFile) {
			os.Remove(dbFile)
		}
	}()
	assert.False(t, InTransaction(db))
	db = db.Begin()
	assert.True(t, InTransaction(db))
	fmt.Println(InTransaction(nil))
}
