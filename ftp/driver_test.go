//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package ftp

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/bigfile/bigfile/databases/models"
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
