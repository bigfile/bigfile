//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package ftp

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"goftp.io/server"
)

func newConn(user string) *server.Conn {
	conn := new(server.Conn)
	connUserAddr := reflect.ValueOf(conn).Elem().FieldByName("user").UnsafeAddr()
	connUserAddrPt := (*string)(unsafe.Pointer(connUserAddr))
	*connUserAddrPt = user
	return conn
}

func TestDriver_Init(t *testing.T) {
	driver := &Driver{}
	driver.Init(newConn("hello"))
	assert.Equal(t, "hello", driver.conn.LoginUser())
}

func TestDriverBuildPath(t *testing.T) {
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)

	assert.Nil(t, trx.Model(token).Update("path", "/test").Error)
	driver := &Driver{db: trx, conn: newConn(tokenPrefix + token.UID)}
	assert.Equal(t, "/test/save/to", driver.buildPath("/save/to"))

	driver = &Driver{db: trx, conn: newConn(token.App.UID)}
	assert.Equal(t, "/save/to", driver.buildPath("/save/to"))
}

func newDriverForTest(t *testing.T) (driver *Driver, down func(*testing.T), err error) {
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	return &Driver{db: trx, app: app, rootPath: &appRootPath}, down, nil
}

func TestDriver_Stat(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	tempDir := models.NewTempDirForTest()
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	_, err = driver.Stat("/not/exists/file.bytes")
	assert.NotNil(t, err)
	assert.True(t, gorm.IsRecordNotFoundError(err))

	_, err = models.CreateOrGetLastDirectory(driver.app, "/this/is/a/dir", driver.db)
	assert.Nil(t, err)
	dirInfo, err := driver.Stat("/this/is/a/dir")
	assert.Nil(t, err)
	assert.True(t, dirInfo.IsDir())
	assert.Equal(t, "dir", dirInfo.Name())

	fileReader := bytes.NewReader(models.Random(222))
	_, err = models.CreateFileFromReader(driver.app, "/this/is/a/dir/file.bytes", fileReader, models.Hidden, &tempDir, driver.db)
	assert.Nil(t, err)
	fileInfo, err := driver.Stat("/this/is/a/dir/file.bytes")
	assert.Nil(t, err)
	assert.False(t, fileInfo.IsDir())
	assert.Equal(t, "file.bytes", fileInfo.Name())
	assert.Equal(t, int64(222), fileInfo.Size())
}

func TestDriver_ChangeDir(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	defer func() { down(t) }()
	assert.Nil(t, driver.ChangeDir("/test/change/dir"))
}

func TestDriver_ListDir(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	tempDir := models.NewTempDirForTest()
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	for index := 0; index < 20; index++ {
		_, err = models.CreateOrGetLastDirectory(driver.app, "/create/dir/"+strconv.Itoa(index), driver.db)
		assert.Nil(t, err)
	}
	_, err = models.CreateFileFromReader(
		driver.app, "/create/dir/file.bytes", strings.NewReader(""), models.Hidden, &tempDir, driver.db)
	assert.Nil(t, err)

	var fileNum = 0
	var dirNum = 0

	assert.Nil(t, driver.ListDir("/create/dir", func(info server.FileInfo) error {
		if info.IsDir() {
			dirNum++
		} else {
			fileNum++
		}
		return nil
	}))
	assert.Equal(t, 20, dirNum)
	assert.Equal(t, 1, fileNum)
}

func TestDriver_DeleteDir(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	defer func() { down(t) }()

	err = driver.DeleteDir("/not/exist")
	assert.NotNil(t, err)
	assert.True(t, gorm.IsRecordNotFoundError(err))

	_, err = models.CreateOrGetLastDirectory(driver.app, "/path/to/dir", driver.db)
	assert.Nil(t, err)
	assert.Nil(t, driver.DeleteDir("/path/to/dir"))
}

func TestDriver_DeleteFile(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	tempDir := models.NewTempDirForTest()
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	err = driver.DeleteFile("/create/dir/file.bytes")
	assert.NotNil(t, err)
	assert.True(t, gorm.IsRecordNotFoundError(err))

	_, err = models.CreateFileFromReader(
		driver.app, "/create/dir/file.bytes", strings.NewReader(""), models.Hidden, &tempDir, driver.db)
	assert.Nil(t, err)
	assert.Nil(t, driver.DeleteFile("/create/dir/file.bytes"))
}

func TestDriver_Rename(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	tempDir := models.NewTempDirForTest()
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()

	err = driver.Rename("/create/dir/file.bytes", "/create/dir/random.bytes")
	assert.NotNil(t, err)
	assert.True(t, gorm.IsRecordNotFoundError(err))

	_, err = models.CreateFileFromReader(
		driver.app, "/create/dir/file.bytes", strings.NewReader(""), models.Hidden, &tempDir, driver.db)
	assert.Nil(t, err)
	assert.Nil(t, driver.Rename("/create/dir/file.bytes", "/create/dir/random.bytes"))
}

func TestDriver_MakeDir(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	defer func() { down(t) }()
	assert.Nil(t, driver.MakeDir("/create/a/directory"))
}

func TestDriver_PutFile(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	tempDir := models.NewTempDirForTest()
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	driver.rootChunkPath = &tempDir

	// append to not exist file
	_, err = driver.PutFile("/save/to/file.bytes", strings.NewReader(""), true)
	assert.NotNil(t, err)
	assert.True(t, gorm.IsRecordNotFoundError(err))

	// append to existed file
	_, err = models.CreateFileFromReader(
		driver.app, "/create/dir/file.bytes", strings.NewReader(""), models.Hidden, &tempDir, driver.db)
	assert.Nil(t, err)
	writeBytes, err := driver.PutFile("/create/dir/file.bytes", bytes.NewReader(models.Random(22)), true)
	assert.Nil(t, err)
	assert.Equal(t, int64(22), writeBytes)

	// create new file
	writeBytes, err = driver.PutFile("/create/dir/random.bytes", bytes.NewReader(models.Random(22)), false)
	assert.Nil(t, err)
	assert.Equal(t, int64(22), writeBytes)
}

func TestDriver_GetFile(t *testing.T) {
	driver, down, err := newDriverForTest(t)
	assert.Nil(t, err)
	tempDir := models.NewTempDirForTest()
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	driver.rootChunkPath = &tempDir

	randomBytes := models.Random(256)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	_, err = models.CreateFileFromReader(
		driver.app, "/create/dir/file.bytes", bytes.NewReader(randomBytes), models.Hidden, &tempDir, driver.db)
	assert.Nil(t, err)

	size, rc, err := driver.GetFile("/create/dir/file.bytes", 0)
	assert.Nil(t, err)
	assert.Equal(t, 256, int(size))
	allContent, err := ioutil.ReadAll(rc)
	assert.Nil(t, err)
	allContentHash, err := util.Sha256Hash2String(allContent)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, allContentHash)
}
