//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"reflect"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var ErrInvalidSortTypes = errors.New("invalid sort types, only one of type, -type, name, -name, time and -time")

type directoryListInput struct {
	Token  string  `form:"token" binding:"required"`
	Nonce  string  `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	SubDir *string `form:"subDir,default=/" binding:"omitempty"`
	Sort   *string `form:"sort,default=-type" binding:"omitempty"`
	Limit  *int    `form:"limit,default=10" binding:"omitempty,min=10,max=20"`
	Offset *int    `form:"offset,default=0" binding:"omitempty,min=0"`
}

// DirectoryList is used to list a directory
func DirectoryListHandler(ctx *gin.Context) {
	var (
		ip                    = ctx.ClientIP()
		db                    = ctx.MustGet("db").(*gorm.DB)
		err                   error
		token                 = ctx.MustGet("token").(*models.Token)
		input                 = ctx.MustGet("inputParam").(*directoryListInput)
		directoryListSrv      *service.DirectoryList
		directoryListSrvValue interface{}
		directoryListSrvResp  *service.DirectoryListResponse

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if err = validateSort(*input.Sort); err != nil {
		generateErrors(err, "sort")
		return
	}

	directoryListSrv = &service.DirectoryList{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		IP:          &ip,
		SubDir:      *input.SubDir,
		Sort:        *input.Sort,
		Offset:      *input.Offset,
		Limit:       *input.Limit,
	}

	if err = directoryListSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if directoryListSrvValue, err = directoryListSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	directoryListSrvResp = directoryListSrvValue.(*service.DirectoryListResponse)

	result := map[string]interface{}{
		"total": directoryListSrvResp.Total,
		"pages": directoryListSrvResp.Pages,
	}

	items := make([]map[string]interface{}, len(directoryListSrvResp.Files))
	for index, item := range directoryListSrvResp.Files {
		if items[index], err = fileResp(&item, db); err != nil {
			reErrors = generateErrors(err, "")
			return
		}
	}

	result["items"] = items
	data = result
	code = 200
}

var preDefinedSortTypes = []string{"type", "-type", "name", "-name", "time", "-time"}

func validateSort(sort string) error {
	for _, s := range preDefinedSortTypes {
		if s == sort {
			return nil
		}
	}
	return ErrInvalidSortTypes
}
