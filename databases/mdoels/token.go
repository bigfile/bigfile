//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"strings"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/jinzhu/gorm"
)

// Token acts as a key, it limits the user that own this token
// which directories can be accessed. Or only when it's used with
// specify ip, it will be accepted. Or some tokens only can be used
// to read file. every token has an expired time, expired token can't
// be used to do anything.
type Token struct {
	ID             uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key" json:"-"`
	UID            string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid" json:"token"`
	Secret         *string    `gorm:"type:CHAR(32)" json:"-"`
	AppID          uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:appId" json:"-"`
	IP             *string    `gorm:"type:VARCHAR(1500);column:ip" json:"ip"`
	AvailableTimes int        `gorm:"type:int(10);column:availableTimes;DEFAULT:-1" json:"availableTimes"`
	ReadOnly       int8       `gorm:"type:tinyint;column:readOnly;DEFAULT:0" json:"readOnly"`
	Path           string     `gorm:"type:tinyint;column:path" json:"path"`
	ExpiredAt      *time.Time `gorm:"type:TIMESTAMP;column:expiredAt" json:"expiredAt"`
	CreatedAt      time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt" json:"-"`
	UpdatedAt      time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt" json:"-"`
	DeletedAt      *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt" json:"-"`

	App App `gorm:"association_foreignkey:id;foreignkey:AppID" json:"-"`
}

// TableName represent token table name
func (t *Token) TableName() string {
	return "tokens"
}

// Scope represent token scope. Actually, it's equal to path.
func (t *Token) Scope() string {
	return t.Path
}

// BeforeSave will be called before token saved
func (t *Token) BeforeSave() (err error) {
	if !strings.HasPrefix(t.Path, "/") {
		t.Path = "/" + t.Path
	}
	return nil
}

// AllowIPAccess is used to check whether this ip can be allowed
// to use this token
func (t *Token) AllowIPAccess(ip string) bool {
	if t.IP == nil {
		return true
	}
	return strings.Contains(*t.IP, ip)
}

// NewToken will generate a token by input params
func NewToken(
	app *App, path string, expiredAt *time.Time, ip, secret *string, availableTimes int, readOnly int8, db *gorm.DB,
) (*Token, error) {
	var (
		token = &Token{
			UID:            bson.NewObjectId().Hex(),
			Secret:         secret,
			AppID:          app.ID,
			IP:             ip,
			AvailableTimes: availableTimes,
			ReadOnly:       readOnly,
			Path:           path,
			ExpiredAt:      expiredAt,
			App:            *app,
		}
		err error
	)
	err = db.Create(token).Error
	return token, err
}

func findTokenByUID(uid string, trashed bool, db *gorm.DB) (*Token, error) {
	var (
		token = &Token{}
		err   error
	)
	if trashed {
		db = db.Unscoped()
	}
	err = db.Preload("App").Where("uid = ?", uid).Find(token).Error
	if err != nil {
		return token, err
	}
	return token, nil
}

// FindTokenByUID find a token by uid
func FindTokenByUID(uid string, db *gorm.DB) (*Token, error) {
	return findTokenByUID(uid, false, db)
}

// FindTokenByUIDWithTrashed find a token by uid with trashed
func FindTokenByUIDWithTrashed(uid string, db *gorm.DB) (*Token, error) {
	return findTokenByUID(uid, true, db)
}
