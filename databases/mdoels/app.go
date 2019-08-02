//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package models map program entity to database table
package models

import (
	"time"
)

// App represent an application in system
type App struct {
	ID        uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Secret    string     `gorm:"type:CHAR(32) NOT NULL"`
	Name      string     `gorm:"type:VARCHAR(100) NOT NULL"`
	Note      *string    `gorm:"type:VARCHAR(500) NULL"`
	CreatedAt time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `gorm:"type:TIMESTAMP(6);INDEX"`
}

// TableName represent table name
func (app *App) TableName() string {
	return "apps"
}
