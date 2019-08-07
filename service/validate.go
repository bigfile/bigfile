//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"errors"
	"regexp"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

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

// ValidatePath is used to validate whether the given path is legal
func ValidatePath(path string) bool {
	var (
		regexps = []*regexp.Regexp{
			regexp.MustCompile(`^(?:/[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255})+$`),
			regexp.MustCompile(`^(?:/[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255})+/$`),
			regexp.MustCompile(`^(?:[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255}/|$)+$?`),
			regexp.MustCompile(`^[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255}$`),
			regexp.MustCompile(`^/$`),
		}
	)

	for _, regex := range regexps {
		if regex.MatchString(path) {
			return true
		}
	}

	return false
}
