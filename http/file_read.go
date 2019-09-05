//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type fileReadInput struct {
	Token         string  `form:"token" binding:"required"`
	FileUID       string  `form:"fileUid" binding:"required"`
	Nonce         *string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	Sign          *string `form:"sign" binding:"omitempty"`
	OpenInBrowser bool    `form:"openInBrowser,default=0" binding:"omitempty"`
}

// FileReadHandler is used to handle file read request
func FileReadHandler(ctx *gin.Context) {
	var (
		ip                     = ctx.ClientIP()
		db                     = ctx.MustGet("db").(*gorm.DB)
		err                    error
		file                   *models.File
		token                  = ctx.MustGet("token").(*models.Token)
		input                  = ctx.MustGet("inputParam").(*fileReadInput)
		requestID              = ctx.GetInt64("requestId")
		fileReadSrv            *service.FileRead
		fileReadSrvValue       interface{}
		fileReadSrvValueReader io.ReadSeeker
	)

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, "fileUid"),
		})
		return
	}

	fileReadSrv = &service.FileRead{
		BaseService: service.BaseService{
			DB: db,
		},
		Token: token,
		File:  file,
		IP:    &ip,
	}

	if isTesting {
		fileReadSrv.RootPath = testingChunkRootPath
	}

	if err = fileReadSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}

	if fileReadSrvValue, err = fileReadSrv.Execute(context.Background()); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}
	fileReadSrvValueReader = fileReadSrvValue.(io.ReadSeeker)
	readAllContents(ctx, fileReadSrvValueReader, file, input)
}

func readAllContents(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *fileReadInput) {
	headers := map[string]string{
		"ETag":                file.Object.Hash,
		"Accept-Ranges":       "bytes",
		"Content-Type":        "application/octet-stream",
		"Last-Modified":       file.UpdatedAt.Format(time.RFC1123),
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, file.Name),
	}
	headers["Content-Length"] = strconv.Itoa(file.Size)
	if contentType := mime.TypeByExtension(path.Ext(file.Name)); contentType != "" {
		headers["Content-Type"] = contentType
	}
	if input.OpenInBrowser {
		headers["Content-Disposition"] = fmt.Sprintf(`inline; filename="%s"`, file.Name)
	}
	ctx.Set("ignoreRespBody", true)
	ctx.DataFromReader(http.StatusOK, int64(file.Size), headers["Content-Type"], readerSeeker, headers)
}
