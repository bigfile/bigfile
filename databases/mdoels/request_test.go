//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestWithProtocol(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	req, err := NewRequestWithProtocol("http", trx)
	assert.Nil(t, err)
	assert.True(t, req.ID > 0)
}

func TestMustNewRequestWithProtocol(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	req := MustNewRequestWithProtocol("http", trx)
	assert.True(t, req.ID > 0)
}

func TestMustNewRequestWithProtocol2(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	MustNewRequestWithProtocol("http", nil)
}

func TestRequest_Save(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	req := MustNewRequestWithProtocol("http", trx)
	assert.True(t, req.ID > 0)

	ip := "192.168.0.1"
	req.IP = &ip
	assert.Nil(t, req.Save(trx))
}

func TestMustNewRequestWithHTTPProtocol(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	req := MustNewRequestWithHTTPProtocol("192.168.0.1", "POST", "/api/bigfile/token/create", trx)
	assert.True(t, req.ID > 0)
}

func TestMustNewRequestWithHTTPProtocol2(t *testing.T) {
	defer func() {
		assert.NotNil(t, recover())
	}()
	MustNewRequestWithHTTPProtocol("192.168.0.1", "POST", "/api/bigfile/token/create", nil)
}
