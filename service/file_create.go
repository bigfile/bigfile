//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"gopkg.in/go-playground/validator.v9"
)

var (
	// ErrPathExisted represent the path has already existed
	ErrPathExisted = errors.New("the path has already existed")
)

// FileCreate is used to upload file or create directory
type FileCreate struct {
	BaseService

	Token     *models.Token `validate:"required"`
	Path      string        `validate:"required,max=1000"`
	Hidden    int8          `validate:"oneof=0 1"`
	IP        *string       `validate:"omitempty"`
	Reader    io.Reader     `validate:"omitempty"`
	Overwrite int8          `validate:"oneof=0 1"`
	Rename    int8          `validate:"oneof=0 1"`
	Append    int8          `validate:"oneof=0 1"`
}

// Validate is used to validate params
func (f *FileCreate) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)

	if errs = Validate.Struct(f); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateToken(f.DB, f.IP, false, f.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileCreate.Token", err))
	}

	if !ValidatePath(f.Path) {
		validateErrors = append(validateErrors, generateErrorByField("FileCreate.Path", ErrInvalidPath))
	}

	return validateErrors
}

// Execute is used to upload file or create directory
func (f *FileCreate) Execute(ctx context.Context) (interface{}, error) {

	var (
		err  error
		path = f.Token.PathWithScope(f.Path)
		file *models.File
	)

	f.BaseService.After = append(f.BaseService.After, func(ctx context.Context, service Service) error {
		f := service.(*FileCreate)
		return f.Token.UpdateAvailableTimes(-1, f.DB)
	})

	if err = f.CallBefore(ctx, f); err != nil {
		return nil, err
	}

	if f.Reader == nil {
		return models.CreateOrGetLastDirectory(&f.Token.App, path, f.DB)
	}

	if file, err = models.FindFileByPath(&f.Token.App, path, f.DB); err != nil && !util.IsRecordNotFound(err) {
		return nil, err
	}

	if file == nil || file.ID == 0 {
		return models.CreateFileFromReader(&f.Token.App, path, f.Reader, f.Hidden, f.RootPath, f.DB)
	}

	if f.Overwrite == 1 {
		return file, file.OverWriteFromReader(f.Reader, f.Hidden, f.RootPath, f.DB)
	}

	if f.Append == 1 {
		return file, file.AppendFromReader(f.Reader, f.Hidden, f.RootPath, f.DB)
	}

	if f.Rename == 1 {
		var (
			dir      = filepath.Dir(path)
			basename = filepath.Base(path)
		)
		path = fmt.Sprintf("%s/%s_%s", dir, models.RandomWithMd5(256), basename)
		return models.CreateFileFromReader(&f.Token.App, path, f.Reader, f.Hidden, f.RootPath, f.DB)
	}

	if f.CallAfter(ctx, f) != nil {
		return nil, err
	}

	return nil, ErrPathExisted
}
