//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/go-playground/validator.v9"

	models "github.com/bigfile/bigfile/databases/mdoels"
)

// TokenCreate provide service for create token
type TokenCreate struct {
	BaseService

	App            *models.App `validate:"required"`
	Path           string      `validate:"required"`
	ExpiredAt      *time.Time  `validate:"omitempty,gt"`
	Ip             *string     `validate:"omitempty,max=1500"`
	Secret         *string     `validate:"omitempty,len=32"`
	AvailableTimes int         `validate:"omitempty,gte=-1"`
	ReadOnly       int8        `validate:"oneof=0 1"`
}

// Validate is used to validate input params
func (c *TokenCreate) Validate() error {

	if errs := Validate.Struct(c); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			if srvErr, ok := ErrMap[err.Namespace()]; ok {
				return srvErr
			} else {
				ErrUnknown.Msg = fmt.Sprintf(ValidateEreTmp, err.Namespace(), err.Field(), err.Tag())
				return ErrUnknown
			}
		}
	}

	if err := ValidateApp(c.DB, c.App); err != nil {
		ErrApp.Desc = err.Error()
		return ErrApp
	}

	return nil
}

// Execute is used to implement token create
func (c *TokenCreate) Execute(ctx context.Context) error {
	return nil
}
