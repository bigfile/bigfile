//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"io"

	"github.com/bigfile/bigfile/databases/models"
	"gopkg.in/go-playground/validator.v9"
)

type FileRead struct {
	BaseService

	Token *models.Token `validate:"required"`
	File  *models.File  `validate:"required"`
	IP    *string       `validate:"omitempty"`
}

// Validate is used to validate service params
func (fr *FileRead) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)
	if errs = Validate.Struct(fr); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateToken(fr.DB, fr.IP, true, fr.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileRead.Token", err))
	}

	if err := ValidateFile(fr.DB, fr.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileRead.File", err))
	} else {
		if err := fr.File.CanBeAccessedByToken(fr.Token, fr.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("FileRead.Token", err))
		}
	}

	return validateErrors
}

// Execute is used to read file
func (fr *FileRead) Execute(ctx context.Context) (interface{}, error) {
	var (
		err        error
		fileReader io.Reader
	)

	fr.BaseService.Before = append(fr.BaseService.After, func(ctx context.Context, service Service) error {
		f := service.(*FileRead)
		return f.Token.UpdateAvailableTimes(-1, f.DB)
	})

	if fileReader, err = fr.File.Reader(fr.RootPath, fr.DB); err != nil {
		return nil, err
	}

	return fileReader, nil
}
