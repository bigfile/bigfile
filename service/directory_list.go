//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

// ErrListFile represent that you are listing a directory
var ErrListFile = errors.New("can't list the content of a file")

// DirectoryListResponse represent the response value of DirectoryList service
type DirectoryListResponse struct {
	Total int
	Pages int
	Files []models.File
}

// DirectoryList is used to list all files and sub directories
type DirectoryList struct {
	BaseService

	Token  *models.Token `validate:"required"`
	IP     *string       `validate:"omitempty"`
	SubDir string        `validate:"omitempty"`
	Sort   string        `validate:"required,oneof=type -type name -name time -time"`
	Offset int           `validate:"omitempty,min=0"`
	Limit  int           `validate:"required,min=10,max=20"`
}

// Validate is used to validate params
func (dl *DirectoryList) Validate() ValidateErrors {
	var (
		err            error
		validateErrors ValidateErrors
	)

	if err = Validate.Struct(dl); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err = ValidateToken(dl.DB, dl.IP, false, dl.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("DirectoryList.Token", err))
	}

	if !ValidatePath(dl.SubDir) {
		validateErrors = append(validateErrors, generateErrorByField("DirectoryList.SubDir", ErrInvalidPath))
	}

	return validateErrors
}

// Execute is used to list a directory
func (dl *DirectoryList) Execute(ctx context.Context) (interface{}, error) {
	var (
		err     error
		dir     *models.File
		total   int
		pages   int
		dirPath = dl.Token.PathWithScope(dl.SubDir)
	)

	if err = dl.Token.UpdateAvailableTimes(-1, dl.DB); err != nil {
		return nil, err
	}

	if dir, err = models.FindFileByPath(&dl.Token.App, dirPath, dl.DB, false); err != nil {
		return nil, err
	}

	if dir.IsDir == 0 {
		return nil, ErrListFile
	}

	total = dl.DB.Model(dir).Association("Children").Count()
	pages = int(math.Ceil(float64(total) / float64(dl.Limit)))

	if err = dl.DB.Preload("Children", func(db *gorm.DB) *gorm.DB {
		var (
			order = "DESC"
			key   = "isDir"
		)
		if !strings.HasPrefix(dl.Sort, "-") {
			order = "ASC"
		}
		switch strings.TrimPrefix(dl.Sort, "-") {
		case "type":
			key = "isDir"
		case "name":
			key = "name"
		case "time":
			key = "updatedAt"
		}
		return db.Order(key + " " + order).Offset(dl.Offset).Limit(dl.Limit)
	}).First(dir).Error; err != nil {
		return nil, err
	}

	return &DirectoryListResponse{
		Total: total,
		Pages: pages,
		Files: dir.Children,
	}, nil
}
