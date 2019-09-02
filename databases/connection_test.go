//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package databases

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConnection(t *testing.T) {
	dbFile, err := ioutil.TempFile(os.TempDir(), "*.db")
	assert.Nil(t, err)
	defer os.Remove(dbFile.Name())

	dbConfig := &config.Database{
		Driver: "sqlite3",
		DBFile: dbFile.Name(),
	}
	_, err = NewConnection(dbConfig)
	assert.Nil(t, err)

	_, err = NewConnection(nil)
	assert.Nil(t, err)

	connection = nil
	dbConfig.Driver = "sqlite"
	_, err = NewConnection(dbConfig)
	assert.NotNil(t, err)
}

func TestMustNewConnection(t *testing.T) {
	defer func() { assert.Nil(t, recover()) }()
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

	connection = nil
	defer func() { assert.NotNil(t, recover()) }()
	MustNewConnection(&config.Database{})
}
