//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

// Package ftp implement ftp protocol for file transfer
package ftp

import (
	"errors"
	"strings"

	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/jinzhu/gorm"
)

var (
	testDbConn *gorm.DB
	// ErrTokenPassword represent the password of token is wrong
	ErrTokenPassword = errors.New("toke password validate failed")

	// ErrAppPassword represent the password of app is wrong
	ErrAppPassword = errors.New("app password validate failed")

	// ErrTokenNotFound represent that wrong token is being used
	ErrTokenNotFound = errors.New("token not found")

	tokenPrefix = "token:"
)

// Auth is used to implement ftp Auth interface
type Auth struct{}

// CheckPasswd is used to check the user whether is correct
func (a *Auth) CheckPasswd(name, password string) (correct bool, err error) {
	var (
		db    *gorm.DB
		app   = &models.App{}
		token *models.Token
	)

	if testDbConn != nil {
		db = testDbConn
	} else {
		db = databases.MustNewConnection(nil)
	}

	if strings.HasPrefix(name, tokenPrefix) {
		tokenUID := strings.TrimPrefix(name, tokenPrefix)
		if token, err = models.FindTokenByUID(tokenUID, db); err != nil {
			return false, ErrTokenNotFound
		}
		if token.Secret != nil && password != *token.Secret {
			return correct, ErrTokenPassword
		}
		return true, nil
	}
	if err = db.Where("uid = ? and secret = ?", name, password).First(app).Error; err != nil {
		return false, ErrAppPassword
	}
	return true, nil
}
