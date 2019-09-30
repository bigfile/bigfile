//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// JpegContentType represent the target content type for http response
const JpegContentType = "image/jpeg"

// ImageConvertInput is a collection of params of request that is from image convert
type ImageConvertInput struct {
	Token         string  `form:"token" binding:"required"`
	FileUID       string  `form:"fileUid" binding:"required"`
	Type          string  `form:"type,default=zoom" binding:"omitempty"`
	Width         float64 `form:"width" binding:"required"`
	Height        float64 `form:"height" binding:"required"`
	Left          int     `form:"left,default=0" binding:"omitempty"`
	Top           int     `form:"top,default=0" binding:"omitempty"`
	Nonce         *string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	Sign          *string `form:"sign" binding:"omitempty"`
	OpenInBrowser bool    `form:"openInBrowser,default=0" binding:"omitempty"`
}

// ImageConvertHandler is used to handle image convert request
func ImageConvertHandler(ctx *gin.Context) {
	var (
		ip              = ctx.ClientIP()
		db              = ctx.MustGet("db").(*gorm.DB)
		err             error
		file            *models.File
		token           = ctx.MustGet("token").(*models.Token)
		input           = ctx.MustGet("inputParam").(*ImageConvertInput)
		requestID       = ctx.GetInt64("requestId")
		imageConvertSrv *service.ImageConvert
		convertData     []byte
	)

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, "fileUid"),
		})
		return
	}

	imageConvertSrv = &service.ImageConvert{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		File:        file,
		IP:          &ip,
		Type:        input.Type,
		Width:       input.Width,
		Height:      input.Height,
		Left:        input.Left,
		Top:         input.Top,
	}

	if isTesting {
		imageConvertSrv.RootPath = testingChunkRootPath
	}

	if err = imageConvertSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}

	if convertData, err = imageConvertSrv.Execute(context.Background()); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}

	ctx.Header("Last-Modified", file.UpdatedAt.Format(time.RFC1123))
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))

	if input.OpenInBrowser {
		ctx.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Name))
	}
	ctx.Set("ignoreRespBody", true)

	ctx.Data(http.StatusOK, JpegContentType, convertData)
}
