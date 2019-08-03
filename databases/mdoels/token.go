//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import "time"

type Token struct {
	ID             uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID            string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	Secret         *string    `gorm:"type:CHAR(32)"`
	AppID          uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:appId"`
	IP             *string    `gorm:"type:VARCHAR(1500);column:ip"`
	AvailableTimes int        `gorm:"type:int(10);column:availableTimes;DEFAULT:-1"`
	ReadOnly       int8       `gorm:"type:tinyint;column:readOnly;DEFAULT:0"`
	Path           string     `gorm:"type:tinyint;column:path"`
	ExpiredAt      *time.Time `gorm:"type:TIMESTAMP;column:expiredAt"`
	CreatedAt      time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt      time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt      *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`

	App App `gorm:"association_foreignkey:id;foreignkey:AppID"`
}

// TableName represent token table name
func (t *Token) TableName() string {
	return "tokens"
}

// Scope represent token scope
func (t *Token) Scope() string {
	return t.Path
}
