//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"time"

	"gopkg.in/go-playground/validator.v9"

	models "github.com/bigfile/bigfile/databases/mdoels"
)

// TokenCreate provide service for create token
type TokenCreate struct {
	BaseService

	Ip             *string     `validate:"omitempty,max=1500"`
	App            *models.App `validate:"required"`
	Path           string      `validate:"required,max=1000"`
	Secret         *string     `validate:"omitempty,len=32"`
	ReadOnly       int8        `validate:"oneof=0 1"`
	ExpiredAt      *time.Time  `validate:"omitempty,gt"`
	AvailableTimes int         `validate:"omitempty,gte=-1"`
}

// Validate is used to validate input params
func (c *TokenCreate) Validate() ValidateErrors {

	var (
		validateErrors ValidateErrors
		errs           error
	)

	if errs = Validate.Struct(c); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateApp(c.DB, c.App); err != nil {
		var (
			field  = "TokenCreate.App"
			appErr = &ValidateError{
				Code:      PreDefinedValidateErrors[field].Code,
				Field:     field,
				Exception: err,
			}
		)
		validateErrors = append(validateErrors, appErr)

	}

	return validateErrors
}

// Execute is used to implement token create
func (c *TokenCreate) Execute(ctx context.Context) error {
	return nil
}
