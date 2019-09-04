//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/bigfile/bigfile/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectReader(t *testing.T) {

	_, err := NewObjectReader(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, ErrInvalidObject)

	db, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)

	object := &Object{}
	assert.Nil(t, db.Save(object).Error)

	_, err = NewObjectReader(object, nil, db)
	assert.NotNil(t, err)
	assert.Equal(t, err, ErrObjectNoChunks)
}

func newObjectForObjectReaderTest(t *testing.T) (*Object, *string, func(*testing.T), *gorm.DB) {
	var (
		err               error
		object            *Object
		tempDir           = NewTempDirForTest()
		randomBytes       = Random(ChunkSize*2 + 123)
		randomBytesHash   string
		randomBytesReader = bytes.NewReader(randomBytes)
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	randomBytesHash, err = util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	object, err = CreateObjectFromReader(randomBytesReader, &tempDir, trx)
	assert.Nil(t, err)
	assert.Equal(t, object.Hash, randomBytesHash)
	assert.Nil(t, trx.Preload("Chunks").Find(object).Error)
	return object, &tempDir, func(t *testing.T) {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}, trx
}

func TestNewObjectReader2(t *testing.T) {
	object, rootPath, down, trx := newObjectForObjectReaderTest(t)
	defer down(t)
	_, err := NewObjectReader(object, rootPath, trx)
	assert.Nil(t, err)
}

func TestObjectReader_Read(t *testing.T) {
	object, rootPath, down, trx := newObjectForObjectReaderTest(t)
	defer down(t)
	or, err := NewObjectReader(object, rootPath, trx)
	assert.Nil(t, err)

	allContent, err := ioutil.ReadAll(or)
	assert.Nil(t, err)
	allContentHash, err := util.Sha256Hash2String(allContent)
	assert.Nil(t, err)
	assert.Equal(t, allContentHash, object.Hash)
}
