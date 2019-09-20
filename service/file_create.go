//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	libPath "path"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"gopkg.in/go-playground/validator.v9"
)

var (
	// ErrPathExisted represent the path has already existed
	ErrPathExisted = errors.New("the path has already existed")
	// ErrOnlyOneRenameAppendOverWrite represent uncertain operation
	ErrOnlyOneRenameAppendOverWrite = errors.New("only one of rename, append and overwrite is allowed")

	// ErrFileHasBeenDeleted represent that the file has been deleted
	ErrFileHasBeenDeleted = errors.New("the file has been deleted")
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
func (fc *FileCreate) Validate() ValidateErrors {
	var (
		err            error
		validateErrors ValidateErrors
	)

	if fc.Overwrite+fc.Rename+fc.Append > 1 {
		validateErrors = append(
			validateErrors,
			generateErrorByField("FileCreate.Operate", ErrOnlyOneRenameAppendOverWrite),
		)
	}

	if err = Validate.Struct(fc); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err = ValidateToken(fc.DB, fc.IP, false, fc.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileCreate.Token", err))
	}

	if !ValidatePath(fc.Path) {
		validateErrors = append(validateErrors, generateErrorByField("FileCreate.Path", ErrInvalidPath))
	}

	return validateErrors
}

// Execute is used to upload file or create directory
func (fc *FileCreate) Execute(ctx context.Context) (interface{}, error) {

	var (
		err   error
		path  = fc.Token.PathWithScope(fc.Path)
		file  *models.File
		inTrx = util.InTransaction(fc.DB)
	)

	if !inTrx {
		fc.DB = fc.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		defer func() {
			if reErr := recover(); reErr != nil {
				fc.DB.Rollback()
			}
		}()
		defer func() { err = fc.DB.Commit().Error }()
	}

	if err = fc.Token.UpdateAvailableTimes(-1, fc.DB); err != nil {
		return nil, err
	}

	if fc.Reader == nil {
		return models.CreateOrGetLastDirectory(&fc.Token.App, path, fc.DB)
	}

	if file, err = models.FindFileByPathWithTrashed(&fc.Token.App, path, fc.DB); err != nil && !util.IsRecordNotFound(err) {
		return nil, err
	}

	if file == nil || file.ID == 0 {
		return models.CreateFileFromReader(&fc.Token.App, path, fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	if file.DeletedAt != nil && (fc.Append == 1 || fc.Overwrite == 1) {
		return nil, ErrFileHasBeenDeleted
	}

	if fc.Overwrite == 1 {
		return file, file.OverWriteFromReader(fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	if fc.Append == 1 {
		return file, file.AppendFromReader(fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	if fc.Rename == 1 {
		var (
			dir      = libPath.Dir(path)
			basename = libPath.Base(path)
		)
		path = fmt.Sprintf("%s/%s_%s", dir, models.RandomWithMD5(256), basename)
		return models.CreateFileFromReader(&fc.Token.App, path, fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	return nil, ErrPathExisted
}
