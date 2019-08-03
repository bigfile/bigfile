//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package databases

import (
	"io/ioutil"
	"os"
	"testing"
	"unsafe"

	"github.com/bigfile/bigfile/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConnection(t *testing.T) {

	dbFile, err := ioutil.TempFile(os.TempDir(), "*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbFile.Name())

	dbConfig := &config.Database{
		Driver: "sqlite3",
		DBFile: dbFile.Name(),
	}

	connection, err := NewConnection(dbConfig, true)
	if err != nil {
		t.Fatal(err)
	}
	connectionPointer := uintptr(unsafe.Pointer(connection))

	if connection, err := NewConnection(dbConfig, true); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, connectionPointer, uintptr(unsafe.Pointer(connection)))
	}

	if connection, err := NewConnection(dbConfig, false); err != nil {
		t.Fatal(err)
	} else {
		assert.NotEqual(t, connectionPointer, uintptr(unsafe.Pointer(connection)))
	}
}

func TestMustNewConnection(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()
	dbFile, err := ioutil.TempFile(os.TempDir(), "*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dbFile.Name())

	dbConfig := &config.Database{
		Driver: "sqlite3",
		DBFile: dbFile.Name(),
	}

	MustNewConnection(dbConfig, true)
}
