//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/gographics/gmagick"
	"github.com/jinzhu/gorm"
)

const JpegContentType = "image/jpeg"

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
		ip               = ctx.ClientIP()
		db               = ctx.MustGet("db").(*gorm.DB)
		err              error
		file             *models.File
		token            = ctx.MustGet("token").(*models.Token)
		input            = ctx.MustGet("inputParam").(*ImageConvertInput)
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
		readAllImage(ctx, fileReaderSeeker, file, input)
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
		readAllImage(ctx, fileReaderSeeker, file, input)
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
			Errors:    generateErrors(ErrWrongHTTPRange, ""),
		})
		return
	}
	readRangeImage(ctx, fileReaderSeeker, file, input, rangeStart, rangeEnd)
}

func readAllImage(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *ImageConvertInput) {
	ctx.Header("Accept-Ranges", "bytes")
	ctx.Header("Last-Modified", file.UpdatedAt.Format(time.RFC1123))
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.Name))

	if input.OpenInBrowser {
		ctx.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Name))
	}
	ctx.Set("ignoreRespBody", true)

	renderImage(ctx, readerSeeker, int64(file.Size), input)
}

func readRangeImage(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *ImageConvertInput, start, end int) {
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
	renderImage(ctx, limitReader, limitSize, input)
}

func renderImage(ctx *gin.Context, reader io.Reader, size int64, input *ImageConvertInput) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(reader, buf); err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}
	ctx.Set("ignoreRespBody", true)
	res, err := imageConvert(buf, input)
	if err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
	}
	ctx.Data(http.StatusPartialContent, JpegContentType, res)
}

func imageConvert(imageBuf []byte, input *ImageConvertInput) ([]byte, error) {
	gm := service.NewGm()
	defer func() {
		gm.MagickWand.Destroy()
		gmagick.Terminate()
	}()
	readErr := gm.MagickWand.ReadImageBlob(imageBuf)
	if readErr != nil {
		return nil, readErr
	}

	switch input.Type {
	case "thumb":
		if err := gm.ImageThumb(input.Width, input.Height); err != nil {
			return nil, err
		}
	case "crop":
		if err := gm.ImageCrop(input.Width, input.Height, input.Left, input.Top); err != nil {
			return nil, err
		}
	case "zoom":
		if err := gm.ImageZoom(input.Width, input.Height); err != nil {
			return nil, err
		}
	}

	if err := gm.MagickWand.SetImageFormat("JPEG"); err != nil {
		return nil, err
	}

	return gm.MagickWand.WriteImageBlob(), nil
}
