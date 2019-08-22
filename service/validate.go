//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"errors"
	"regexp"

	"github.com/bigfile/bigfile/databases/models"
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

// ValidateToken is used to validate whether the token is valid
func ValidateToken(db *gorm.DB, ip *string, canReadOnly bool, token *models.Token) error {
	var err error
	if token == nil {
		return errors.New("invalid token")
	}
	if token, err = models.FindTokenByUID(token.UID, db); err != nil {
		return err
	}

	if ip != nil && !token.AllowIPAccess(*ip) {
		return errors.New("token can't be used by this ip")
	}

	if token.AvailableTimes != -1 && token.AvailableTimes <= 0 {
		return errors.New("the available times of token has already exhausted")
	}

	if !canReadOnly && token.ReadOnly == 1 {
		return errors.New("this token is read only")
	}

	return nil
}

// ValidatePath is used to validate whether the given path is legal
func ValidatePath(path string) bool {
	var (
		regexps = []*regexp.Regexp{
			// different regex match different path, see the test case
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
