//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package models map program entity to database table
package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// App represent an application in system
type App struct {
	ID        uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID       string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	Secret    string     `gorm:"type:CHAR(32) NOT NULL"`
	Name      string     `gorm:"type:VARCHAR(100) NOT NULL"`
	Note      *string    `gorm:"type:VARCHAR(500) NULL"`
	CreatedAt time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`
}

// TableName represent table name
func (app *App) TableName() string {
	return "apps"
}

// AfterCreate hooks will be called automatically after app created
func (app *App) AfterCreate(tx *gorm.DB) error {
	var file = &File{
		UID:   UID(),
		PID:   0,
		AppID: app.ID,
		Name:  "",
		IsDir: 1,
	}
	return tx.Save(file).Error
}

// NewApp generate a new application by name and note
func NewApp(name string, note *string, db *gorm.DB) (*App, error) {
	var (
		app = &App{
			Name:   name,
			Note:   note,
			UID:    UID(),
			Secret: NewSecret(),
		}
		err error
	)
	err = db.Create(app).Error
	return app, err
}

func deleteApp(app *App, soft bool, db *gorm.DB) error {
	if !soft {
		db = db.Unscoped()
	}
	return db.Delete(app).Error
}

// DeleteAppSoft execute a soft delete, just mark app is deleted
func DeleteAppSoft(app *App, db *gorm.DB) error {
	return deleteApp(app, true, db)
}

// DeleteAppPermanently delete app app permanently
func DeleteAppPermanently(app *App, db *gorm.DB) error {
	return deleteApp(app, false, db)
}

func findAppByUID(uid string, trashed bool, db *gorm.DB) (*App, error) {
	var (
		app = &App{}
		err error
	)
	if trashed {
		db = db.Unscoped()
	}
	err = db.Where("uid = ?", uid).Find(app).Error
	if err != nil {
		return app, err
	}
	return app, nil
}

// FindAppByUID find an application by uid
func FindAppByUID(uid string, db *gorm.DB) (*App, error) {
	return findAppByUID(uid, false, db)
}

// FindAppByUIDWithTrashed find an application by uid with trashed
func FindAppByUIDWithTrashed(uid string, db *gorm.DB) (*App, error) {
	return findAppByUID(uid, true, db)
}

// DeleteAppByUIDSoft soft delete an app by uid
func DeleteAppByUIDSoft(uid string, db *gorm.DB) error {
	app, err := FindAppByUID(uid, db)
	if err != nil {
		return err
	}
	return deleteApp(app, true, db)
}

// DeleteAppByUIDPermanently delete an app permanently
func DeleteAppByUIDPermanently(uid string, db *gorm.DB) error {
	app, err := FindAppByUID(uid, db)
	if err != nil {
		return err
	}
	return deleteApp(app, false, db)
}
