//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package databases

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bigfile/bigfile/config"
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

	_, err = NewConnection(dbConfig)
	if err != nil {
		t.Fatal(err)
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

	MustNewConnection(dbConfig)
}
