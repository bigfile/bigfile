//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"context"
	"reflect"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type fileDeleteInput struct {
	Token   string `form:"token" binding:"required"`
	Nonce   string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	FileUID string `form:"fileUid" binding:"required"`
	Force   bool   `form:"force,default=0"  binding:"omitempty"`
}

// FileDeleteHandler is used to delete a file or a directory
func FileDeleteHandler(ctx *gin.Context) {
	var (
		ip                 = ctx.ClientIP()
		db                 = ctx.MustGet("db").(*gorm.DB)
		err                error
		file               *models.File
		token              = ctx.MustGet("token").(*models.Token)
		input              = ctx.MustGet("inputParam").(*fileDeleteInput)
		requestID          = ctx.GetInt64("requestId")
		fileDeleteSrv      *service.FileDelete
		fileDeleteSrvValue interface{}

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: requestID,
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		reErrors = generateErrors(err, "fileUid")
		return
	}

	fileDeleteSrv = &service.FileDelete{
		BaseService: service.BaseService{
			DB: db,
		},
		Token: token,
		File:  file,
		Force: &input.Force,
		IP:    &ip,
	}

	if err = fileDeleteSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "system")
		return
	}

	if fileDeleteSrvValue, err = fileDeleteSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "system")
		return
	}

	if data, err = fileResp(fileDeleteSrvValue.(*models.File), db); err != nil {
		reErrors = generateErrors(err, "system")
		return
	}

	code = 200
}
