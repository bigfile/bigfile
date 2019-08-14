//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"time"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"gopkg.in/go-playground/validator.v9"
)

// TokenCreate provide service for create token
type TokenCreate struct {
	BaseService

	IP             *string     `validate:"omitempty,max=1500"`
	App            *models.App `validate:"required"`
	Path           string      `validate:"required,max=1000"`
	Secret         *string     `validate:"omitempty,len=32"`
	ReadOnly       int8        `validate:"oneof=0 1"`
	ExpiredAt      *time.Time  `validate:"omitempty,gt"`
	AvailableTimes int         `validate:"omitempty,gte=-1,max=2147483647"`

	token *models.Token
}

// Validate is used to validate input params
func (t *TokenCreate) Validate() ValidateErrors {

	var (
		validateErrors ValidateErrors
		errs           error
	)

	if errs = Validate.Struct(t); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateApp(t.DB, t.App); err != nil {
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

	if !ValidatePath(t.Path) {
		var (
			field   = "TokenCreate.Path"
			pathErr = &ValidateError{
				Code:      PreDefinedValidateErrors[field].Code,
				Field:     field,
				Exception: errors.New("path is not a legal unix path"),
			}
		)
		validateErrors = append(validateErrors, pathErr)
	}

	return validateErrors
}

// Execute is used to implement token create
func (t *TokenCreate) Execute(ctx context.Context) (interface{}, error) {
	var err error

	if err = t.CallBefore(ctx, t); err != nil {
		return nil, err
	}

	if t.token, err = models.NewToken(
		t.App, t.Path, t.ExpiredAt, t.IP, t.Secret, t.AvailableTimes, t.ReadOnly, t.DB,
	); err != nil {
		return nil, err
	}

	if t.CallAfter(ctx, t) != nil {
		return t.token, err
	}
	return t.token, nil
}
