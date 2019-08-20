//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import "time"

// File represent a file or a directory of system. If it's a file
// it has to associate with an object. Actually, the object hold
// the real content of file.
type File struct {
	ID            uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID           string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	PID           uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:pid"`
	AppID         uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:appId"`
	ObjectID      uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:objectId"`
	Size          int        `gorm:"type:int;column:size"`
	Name          string     `gorm:"type:VARCHAR(255);NOT NULL;column:name"`
	Ext           string     `gorm:"type:VARCHAR(255);NOT NULL;column:ext"`
	IsDir         int8       `gorm:"type:tinyint;column:isDir;DEFAULT:0"`
	Hidden        int8       `gorm:"type:tinyint;column:hidden;DEFAULT:0"`
	DownloadCount uint64     `gorm:"type:BIGINT(20);column:downloadCount;DEFAULT:0"`
	CreatedAt     time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt     time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt     *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`

	Object Object `gorm:"foreignkey:objectId"`
	App    App    `gorm:"foreignkey:appId"`
	Parent *File  `gorm:"foreignkey:pid"`
}

// TableName represent the name of files table
func (f File) TableName() string {
	return "files"
}
