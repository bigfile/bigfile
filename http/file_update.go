//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"reflect"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type fileUpdateInput struct {
	Token   string  `form:"token" binding:"required"`
	FileUID string  `form:"fileUid" binding:"required"`
	Nonce   string  `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	Sign    *string `form:"sign" binding:"omitempty"`
	Hidden  *int8   `form:"hidden" binding:"omitempty"`
	Path    *string `form:"path" binding:"required,max=1000"`
}

// FileUpdateHandler is used to handle file update request
func FileUpdateHandler(ctx *gin.Context) {
	var (
		ip                 = ctx.ClientIP()
		db                 = ctx.MustGet("db").(*gorm.DB)
		err                error
		file               *models.File
		token              = ctx.MustGet("token").(*models.Token)
		input              = ctx.MustGet("inputParam").(*fileUpdateInput)
		fileUpdateSrv      *service.FileUpdate
		fileUpdateSrvValue interface{}

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

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		reErrors = generateErrors(err, "fileUid")
		return
	}

	fileUpdateSrv = &service.FileUpdate{
		BaseService: service.BaseService{
			DB: db,
		},
		Token:  token,
		File:   file,
		IP:     &ip,
		Hidden: input.Hidden,
		Path:   input.Path,
	}

	if isTesting {
		fileUpdateSrv.RootPath = testingChunkRootPath
	}

	if err = fileUpdateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if fileUpdateSrvValue, err = fileUpdateSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	if data, err = fileResp(fileUpdateSrvValue.(*models.File), db); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	code = 200
	success = true
}
