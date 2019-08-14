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

// TokenUpdate provide service for updating token
type TokenUpdate struct {
	BaseService

	Token          string     `validate:"required"`
	IP             *string    `validate:"omitempty,max=1500"`
	Path           *string    `validate:"omitempty,max=1000"`
	Secret         *string    `validate:"omitempty,len=32"`
	ReadOnly       *int8      `validate:"omitempty,oneof=0 1"`
	ExpiredAt      *time.Time `validate:"omitempty,gt"`
	AvailableTimes *int       `validate:"omitempty,gte=-1,max=2147483647"`
}

// Validate is used to validate input params
func (t *TokenUpdate) Validate() ValidateErrors {

	var (
		validateErrors ValidateErrors
		errs           error
	)

	if errs = Validate.Struct(t); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if t.Path != nil {
		if !ValidatePath(*t.Path) {
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
	}

	return validateErrors
}

// Execute is used to update token
func (t *TokenUpdate) Execute(ctx context.Context) (result interface{}, err error) {

	var (
		token *models.Token
	)

	defer func() {
		if recover() != nil || err != nil {
			t.DB.Rollback()
		}
	}()

	if err = t.CallBefore(ctx, t); err != nil {
		return nil, err
	}

	if token, err = models.FindTokenByUID(t.Token, t.DB); err != nil {
		return nil, err
	}

	if t.Path != nil {
		token.Path = *t.Path
	}
	if t.IP != nil {
		token.IP = t.IP
	}
	if t.Secret != nil {
		token.Secret = t.Secret
	}
	if t.ReadOnly != nil {
		token.ReadOnly = *t.ReadOnly
	}
	if t.ExpiredAt != nil {
		token.ExpiredAt = t.ExpiredAt
	}
	if t.AvailableTimes != nil {
		token.AvailableTimes = *t.AvailableTimes
	}

	if t.DB.Save(token).Error != nil {
		return nil, err
	}

	if t.CallAfter(ctx, t) != nil {
		return nil, err
	}

	return token, err
}
