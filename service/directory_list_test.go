//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package service

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestDirectoryList_Validate(t *testing.T) {
	trx, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	directoryListSrv := &DirectoryList{
		BaseService: BaseService{
			DB: trx,
		},
		Token:  nil,
		IP:     nil,
		SubDir: "/$#%/",
		Sort:   "hello",
		Offset: -1,
		Limit:  0,
	}
	err := directoryListSrv.Validate()
	assert.True(t, err.ContainsErrCode(10031))
	assert.True(t, err.ContainsErrCode(10032))
	assert.True(t, err.ContainsErrCode(10034))
	assert.True(t, err.ContainsErrCode(10035))
	assert.Contains(t, err.Error(), "invalid token")
	assert.Contains(t, err.Error(), "path is not a legal unix path")
}

func TestDirectoryList_Execute(t *testing.T) {
	var (
		ok                       bool
		trx                      *gorm.DB
		err                      error
		down                     func(*testing.T)
		token                    *models.Token
		tempDir                  = models.NewTempDirForTest()
		directoryListSrvValue    interface{}
		directoryListSrvResponse *DirectoryListResponse
	)
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	token.Path = "/save"
	assert.Nil(t, trx.Model(token).Update("path", token.Path).Error)

	directoryListSrv := &DirectoryList{
		BaseService: BaseService{DB: trx},
		Token:       token,
		IP:          nil,
		SubDir:      "",
		Sort:        "-type",
		Offset:      0,
		Limit:       10,
	}
	assert.Nil(t, directoryListSrv.Validate())

	_, err = directoryListSrv.Execute(context.TODO())
	assert.True(t, gorm.IsRecordNotFoundError(err))

	saveDir, err := models.CreateOrGetLastDirectory(&token.App, "/save", trx)
	assert.Nil(t, err)

	directoryListSrvValue, err = directoryListSrv.Execute(context.TODO())
	assert.Nil(t, err)
	directoryListSrvResponse, ok = directoryListSrvValue.(*DirectoryListResponse)
	assert.True(t, ok)
	assert.Equal(t, 0, directoryListSrvResponse.Total)
	assert.Equal(t, 0, directoryListSrvResponse.Pages)
	assert.Equal(t, 0, len(directoryListSrvResponse.Files))

	// create sub directory
	for i := 0; i < 15; i++ {
		_, err = models.CreateOrGetLastDirectory(&token.App, "/save/"+strconv.Itoa(i), trx)
		assert.Nil(t, err)
	}

	// first page
	directoryListSrvValue, err = directoryListSrv.Execute(context.TODO())
	assert.Nil(t, err)
	directoryListSrvResponse, ok = directoryListSrvValue.(*DirectoryListResponse)
	assert.True(t, ok)
	assert.Equal(t, 15, directoryListSrvResponse.Total)
	assert.Equal(t, 2, directoryListSrvResponse.Pages)
	assert.Equal(t, 10, len(directoryListSrvResponse.Files))

	// second page
	directoryListSrv.Offset = 10
	assert.Nil(t, directoryListSrv.Validate())
	directoryListSrvValue, err = directoryListSrv.Execute(context.TODO())
	assert.Nil(t, err)
	directoryListSrvResponse, ok = directoryListSrvValue.(*DirectoryListResponse)
	assert.True(t, ok)
	assert.Equal(t, 15, directoryListSrvResponse.Total)
	assert.Equal(t, 2, directoryListSrvResponse.Pages)
	assert.Equal(t, 5, len(directoryListSrvResponse.Files))

	// sort: name asc
	directoryListSrv.Sort = "name"
	assert.Nil(t, directoryListSrv.Validate())
	_, err = directoryListSrv.Execute(context.TODO())
	assert.Nil(t, err)

	// sort: time asc
	directoryListSrv.Sort = "time"
	assert.Nil(t, directoryListSrv.Validate())
	_, err = directoryListSrv.Execute(context.TODO())
	assert.Nil(t, err)

	// list the content of a file, raise an error
	_, err = models.CreateFileFromReader(&token.App, "/save/1/random.bytes", strings.NewReader(""), int8(0), &tempDir, trx)
	assert.Nil(t, err)
	token.Path = "/save/1/random.bytes"
	assert.Nil(t, trx.Model(token).Update("path", "/save/1/random.bytes").Error)
	assert.Nil(t, directoryListSrv.Validate())
	_, err = directoryListSrv.Execute(context.TODO())
	assert.Equal(t, ErrListFile, err)

	assert.Nil(t, trx.Delete(saveDir).Error)
	token.Path = "/save"
	assert.Nil(t, trx.Model(token).Update("path", "/save").Error)
	assert.Nil(t, directoryListSrv.Validate())
	_, err = directoryListSrv.Execute(context.TODO())
	assert.True(t, gorm.IsRecordNotFoundError(err))
}
