//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ErrWrongRangeHeader represent the http range header format error
var ErrWrongRangeHeader = errors.New("http range header format error")

// ErrWrongHttpRange represent wrong http range header
var ErrWrongHttpRange = errors.New("wrong http range header, start must be less than end")

const binaryContentType = "application/octet-stream"

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
		ip               = ctx.ClientIP()
		db               = ctx.MustGet("db").(*gorm.DB)
		err              error
		file             *models.File
		token            = ctx.MustGet("token").(*models.Token)
		input            = ctx.MustGet("inputParam").(*fileReadInput)
		requestID        = ctx.GetInt64("requestId")
		fileReadSrv      *service.FileRead
		fileReaderSeeker io.ReadSeeker
		fileReadSrvValue interface{}
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
		BaseService: service.BaseService{DB: db},
		Token:       token,
		File:        file,
		IP:          &ip,
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
	fileReaderSeeker = fileReadSrvValue.(io.ReadSeeker)
	rangeHeader := ctx.Request.Header.Get("Range")
	if rangeHeader == "" {
		readAllContent(ctx, fileReaderSeeker, file, input)
		return
	}
	rangeHeaderPattern := regexp.MustCompile(`^bytes=(?P<start>\d*)-(?P<end>\d*)$`)
	if !rangeHeaderPattern.Match([]byte(rangeHeader)) {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(ErrWrongRangeHeader, ""),
		})
		return
	}
	rangePosition := strings.TrimPrefix(rangeHeader, "bytes=")
	rangeStart := 0
	rangeEnd := file.Size
	if rangePosition == "-" {
		readAllContent(ctx, fileReaderSeeker, file, input)
		return
	} else if strings.HasPrefix(rangePosition, "-") {
		rangeEnd, _ = strconv.Atoi(strings.TrimPrefix(rangePosition, "-"))
	} else if strings.HasSuffix(rangePosition, "-") {
		rangeStart, _ = strconv.Atoi(strings.TrimSuffix(rangePosition, "-"))
	} else {
		rangePositionSplit := strings.Split(rangePosition, "-")
		rangeStart, _ = strconv.Atoi(rangePositionSplit[0])
		rangeEnd, _ = strconv.Atoi(rangePositionSplit[1])
	}
	if rangeStart > rangeEnd {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(ErrWrongHttpRange, ""),
		})
		return
	}
	readRangeContent(ctx, fileReaderSeeker, file, input, rangeStart, rangeEnd)
}

func readAllContent(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *fileReadInput) {
	headers := map[string]string{
		"ETag":                file.Object.Hash,
		"Accept-Ranges":       "bytes",
		"Content-Type":        binaryContentType,
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

func readRangeContent(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *fileReadInput, start, end int) {
	if _, err := readerSeeker.Seek(int64(start), io.SeekStart); err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}
	limitSize := int64(end - start)
	limitReader := io.LimitReader(readerSeeker, limitSize)
	ctx.Set("ignoreRespBody", true)
	ctx.DataFromReader(http.StatusPartialContent, limitSize, binaryContentType, limitReader, map[string]string{
		"Content-Range": fmt.Sprintf("%d-%d/%d", start, end, file.Size),
	})
}
