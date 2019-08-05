//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"errors"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

const ValidateEreTmp = "Key: '%s' Error:Field validation for '%s' failed on the '%s' tag"

var (
	// Validate represent a global validator
	Validate = validator.New()
)

// ValidateApp is used to validate whether app is valid
func ValidateApp(db *gorm.DB, app *models.App) error {
	if app == nil {
		return errors.New("invalid application")
	}
	if _, err := models.FindAppByUID(app.UID, db); err != nil {
		return err
	}
	return nil
}
