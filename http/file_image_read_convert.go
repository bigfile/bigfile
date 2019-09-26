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

// ErrWrongRangeHeader represent the http range header format error
//var ErrWrongRangeHeader = errors.New("http range header format error")
//
//// ErrWrongHTTPRange represent wrong http range header
//var ErrWrongHTTPRange = errors.New("wrong http range header, start must be less than end")
//
const JpegContentType = "image/jpeg"

type imageFileReadInput struct {
	Token         string  `form:"token" binding:"required"`
	FileUID       string  `form:"fileUid" binding:"required"`
	Type 		  string  `form:"type,default=zoom" binding:"omitempty"`
	Width         string  `form:"width" binding:"required"`
	Height        string  `form:"height" binding:"required"`
	Left          string  `form:"left" binding:"omitempty"`
	Top           string  `form:"top" binding:"omitempty"`
	Nonce         *string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	Sign          *string `form:"sign" binding:"omitempty"`
	OpenInBrowser bool    `form:"openInBrowser,default=0" binding:"omitempty"`
}

// FileReadHandler is used to handle file read request
func ImageFileReadHandler(ctx *gin.Context) {
	var (
		ip               = ctx.ClientIP()
		db               = ctx.MustGet("db").(*gorm.DB)
		err              error
		file             *models.File
		token            = ctx.MustGet("token").(*models.Token)
		input            = ctx.MustGet("inputParam").(*imageFileReadInput)
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

func readAllImage(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *imageFileReadInput) {
	ctx.Header("Accept-Ranges","bytes")
	ctx.Header("Last-Modified",file.UpdatedAt.Format(time.RFC1123))
	ctx.Header("Content-Disposition",fmt.Sprintf(`attachment; filename="%s"`, file.Name))

	if input.OpenInBrowser {
		ctx.Header("Content-Disposition",fmt.Sprintf(`inline; filename="%s"`, file.Name))
	}
	ctx.Set("ignoreRespBody", true)

	renderImage(ctx, readerSeeker, int64(file.Size), input)
}

func readRangeImage(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *imageFileReadInput, start, end int) {
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

func renderImage(ctx *gin.Context, reader io.Reader, size int64, input *imageFileReadInput) {
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
	res,err :=  ImageHandler(ctx,buf,input)
	if err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
	}
	ctx.Data(http.StatusPartialContent, JpegContentType,res)
}

func imageThumb (mw *gmagick.MagickWand,width ,height float64) {
	w := mw.GetImageWidth()
	h := mw.GetImageHeight()
	x := float64(w)/float64(h)
	var	targetW,targetH uint
	if height*x > width {
		targetW = uint(width)
		targetH = uint(width / x)
	} else {
		targetW = uint(height *x)
		targetH = uint(height)
	}
	mw.ResizeImage(targetW, targetH, gmagick.FILTER_LANCZOS, 1)
}

func imageCrop (mw *gmagick.MagickWand,width ,height uint,left ,top int) {
	mw.CropImage(width ,height ,left ,top)
}

func imageZoom (mw *gmagick.MagickWand,width ,height float64) {
	var left,top int
	var x,xW,xH float64

	w := mw.GetImageWidth()
	h := mw.GetImageHeight()
	xW = float64(w)/width
	xH = float64(h)/height
	if xW < xH {
		x = xW
	} else {
		x = xH
	}
	thumbWidth,thumbHeight := uint(float64(w)/x),uint(float64(h)/x)
	mw.ResizeImage(thumbWidth, thumbHeight, gmagick.FILTER_LANCZOS, 1)

	top = int(thumbHeight - uint(height))/2
	left = int(thumbWidth - uint(width))/2
	mw.CropImage(uint(width) , uint(height) ,left ,top)
}

func toUint(s string) uint {
	i,e := strconv.Atoi(s)
	if e!=nil {
		fmt.Println(e)
	}
	return uint(i)
}

func ImageHandler(ctx *gin.Context,imageBuf []byte,input *imageFileReadInput) ([]byte ,error) {
	mw := gmagick.NewMagickWand()
	gmagick.Initialize()
	defer func() {
		mw.Destroy()
		gmagick.Terminate()
	}()
	readErr := mw.ReadImageBlob(imageBuf)
	if readErr != nil {
		return nil, readErr
	}

	width,_ := strconv.ParseFloat(input.Width,64)
	height,_ := strconv.ParseFloat(input.Height,64)

	switch input.Type {
	case "thumb":
		imageThumb(mw,width,height)
	case "crop":
		cw,ch:= toUint(input.Width),toUint(input.Height)
		cl,_ := strconv.Atoi(input.Left)
		ct,_ := strconv.Atoi(input.Top)
		imageCrop(mw,cw,ch,cl,ct)
	case "zoom":
		imageZoom(mw,width,height)
	}
	mw.SetImageFormat("JPEG")

	return mw.WriteImageBlob(), nil
}