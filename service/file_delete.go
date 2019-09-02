//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"gopkg.in/go-playground/validator.v9"
)

// FileDelete is used to provide service for deleting file
type FileDelete struct {
	BaseService

	Token *models.Token `validate:"required"`
	File  *models.File  `validate:"required"`
	Force *bool         `validate:"omitempty"`
	IP    *string       `validate:"omitempty"`
}

// Validate is used to validate service params
func (fd *FileDelete) Validate() ValidateErrors {
	var (
		err            error
		validateErrors ValidateErrors
	)
	if err = Validate.Struct(fd); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err = ValidateToken(fd.DB, fd.IP, true, fd.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileDelete.Token", err))
	}

	if err = ValidateFile(fd.DB, fd.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileDelete.File", err))
	} else {
		if err = fd.File.CanBeAccessedByToken(fd.Token, fd.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("FileDelete.Token", err))
		}
	}

	return validateErrors
}

// Execute is used to provide file delete service
func (fd *FileDelete) Execute(ctx context.Context) (interface{}, error) {
	var (
		falseValue = false
		err        error
		inTrx      = util.InTransaction(fd.DB)
	)

	if !inTrx {
		fd.DB = fd.DB.Begin()
		defer func() {
			if reErr := recover(); reErr != nil {
				fd.DB.Rollback()
			}
		}()
		defer func() { err = fd.DB.Commit().Error }()
	}

	if err = fd.Token.UpdateAvailableTimes(-1, fd.DB); err != nil {
		return nil, err
	}

	if fd.Force == nil {
		fd.Force = &falseValue
	}

	if err = fd.File.Delete(*fd.Force, fd.DB); err != nil {
		return fd.File, err
	}

	return fd.File, nil
}
