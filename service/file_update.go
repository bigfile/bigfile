//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"gopkg.in/go-playground/validator.v9"
)

// FileUpdate is used uo update a file, such as move file to another path,
// or rename file, hide file.
type FileUpdate struct {
	BaseService

	Token  *models.Token `validate:"required"`
	File   *models.File  `validate:"required"`
	IP     *string       `validate:"omitempty"`
	Hidden *int8         `validate:"omitempty,oneof=0 1"`
	Path   *string       `validate:"omitempty,max=1000"`
}

// Validate is used to validate service params
func (fu *FileUpdate) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)
	if errs = Validate.Struct(fu); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateToken(fu.DB, fu.IP, true, fu.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileUpdate.Token", err))
	}

	if err := ValidateFile(fu.DB, fu.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileUpdate.File", err))
	} else {
		if err := fu.File.CanBeAccessedByToken(fu.Token, fu.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("FileUpdate.Token", err))
		}
	}

	if fu.Path != nil {
		if !ValidatePath(*fu.Path) {
			validateErrors = append(validateErrors, generateErrorByField("FileUpdate.Path", ErrInvalidPath))
		}
	}

	return validateErrors
}

// Execute is used to update file
func (fu *FileUpdate) Execute(ctx context.Context) (interface{}, error) {
	var (
		err   error
		inTrx = util.InTransaction(fu.DB)
	)

	fu.BaseService.Before = append(fu.BaseService.After, func(ctx context.Context, service Service) error {
		fu := service.(*FileUpdate)
		return fu.Token.UpdateAvailableTimes(-1, fu.DB)
	})

	if err = fu.CallBefore(ctx, fu); err != nil {
		return nil, err
	}

	if !inTrx {
		fu.DB = fu.DB.Begin()
		defer func() {
			if reErr := recover(); reErr != nil {
				fu.DB.Rollback()
			} else if err != nil {
				if _, ok := err.(*commitError); !ok {
					fu.DB.Rollback()
				}
			}
		}()
		defer func() {
			if err = fu.DB.Commit().Error; err != nil {
				err = &commitError{err: err}
			}
		}()
	}

	if fu.Path != nil {
		if err := fu.File.MoveTo(fu.Token.PathWithScope(*fu.Path), fu.DB); err != nil {
			return nil, err
		}
	}

	if fu.Hidden != nil {
		fu.File.Hidden = *fu.Hidden
	}

	if err := fu.DB.Save(fu.File).Error; err != nil {
		return nil, err
	}

	if err = fu.CallAfter(ctx, fu); err != nil {
		return fu.File, err
	}

	return fu.File, nil
}
