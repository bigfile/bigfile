//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"testing"

	"github.com/bigfile/bigfile/databases"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	var (
		note = "test app"
		app  *App
		err  error
	)
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	app, err = NewApp("test", &note, trx)
	assert.Equal(t, err, nil)
	assert.Equal(t, true, app.ID > 0)
}

func BenchmarkNewApp(b *testing.B) {
	defer func() {
		if err := recover(); err != nil {
			b.Fatal(err)
		}
	}()
	trx := databases.MustNewConnection(nil).Begin()
	defer func() {
		trx.Rollback()
	}()
	for i := 0; i < b.N; i++ {
		_, err := NewApp("test", nil, trx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestDeleteAppSoft(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	assert.Equal(t, true, app.ID > 0)
	err = DeleteAppSoft(app, trx)
	assert.Equal(t, err, nil)
	appTmp := &App{}
	trx.Where("id = ?", app.ID).Find(appTmp)
	assert.Equal(t, true, appTmp.ID == 0)
	trx.Unscoped().Where("id = ?", app.ID).Find(appTmp)
	assert.Equal(t, appTmp.ID, app.ID)
}

func TestDeleteAppPermanently(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	assert.Equal(t, true, app.ID > 0)
	err = DeleteAppPermanently(app, trx)
	assert.Equal(t, err, nil)
	appTmp := &App{}
	trx.Unscoped().Where("id = ?", app.ID).Find(appTmp)
	assert.Equal(t, appTmp.ID, uint64(0))
}

func TestFindAppByUID(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	appTmp, err := FindAppByUID(app.UID, trx)
	assert.Equal(t, err, nil)
	assert.Equal(t, appTmp.ID, app.ID)
}

func TestFindAppByUID2(t *testing.T) {
	trx, down := setUpTestCaseWithTrx(nil, t)
	defer down(t)
	app, err := FindAppByUID("this is a fake uid", trx)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, uint64(0), app.ID)
	assert.Contains(t, err.Error(), "record not found")
}

func TestDeleteAppByUIDSoft(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	assert.Equal(t, true, app.ID > 0)
	err = DeleteAppByUIDSoft(app.UID, trx)
	assert.Equal(t, err, nil)
	appTmp := &App{}
	trx.Where("id = ?", app.ID).Find(appTmp)
	assert.Equal(t, true, appTmp.ID == 0)
	trx.Unscoped().Where("id = ?", app.ID).Find(appTmp)
	assert.Equal(t, appTmp.ID, app.ID)
}

func TestDeleteAppByUIDPermanently(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	assert.Equal(t, true, app.ID > 0)
	err = DeleteAppByUIDPermanently(app.UID, trx)
	assert.Equal(t, err, nil)
	appTmp := &App{}
	trx.Unscoped().Where("id = ?", app.ID).Find(appTmp)
	assert.Equal(t, appTmp.ID, uint64(0))
}

func TestFindAppByUIDWithTrashed(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)
	assert.Equal(t, true, app.ID > 0)
	err = DeleteAppByUIDSoft(app.UID, trx)
	assert.Equal(t, err, nil)

	_, err = FindAppByUID(app.UID, trx)
	assert.NotEqual(t, nil, err)
	assert.Contains(t, err.Error(), "record not found")

	appTmp, err := FindAppByUIDWithTrashed(app.UID, trx)
	assert.Equal(t, nil, err)
	assert.True(t, appTmp.ID > 0)
}

func TestApp_TableName(t *testing.T) {
	assert.Equal(t, (&App{}).TableName(), "apps")
}

func TestApp_AfterCreate(t *testing.T) {
	app, trx, down, err := newAppForTest(nil, t)
	assert.Equal(t, err, nil)
	defer down(t)

	file := &File{}
	assert.Nil(t, trx.Where("appId = ?", app.ID).Find(file).Error)
	assert.True(t, file.ID > 0)
	assert.Equal(t, file.AppID, app.ID)
	assert.Equal(t, int8(1), file.IsDir)
}
