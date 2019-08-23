//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var testingChunkRootPath *string

type fileCreateInput struct {
	Token     string  `form:"token" binding:"required"`
	Nonce     string  `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Path      string  `form:"path" binding:"required,max=1000"`
	Sign      *string `form:"sign" binding:"omitempty"`
	Hash      *string `form:"hash" binding:"omitempty"`
	Size      *int    `form:"size" binding:"omitempty"`
	Overwrite *bool   `form:"overwrite,default=0" binding:"omitempty"`
	Rename    *bool   `form:"rename,default=0" binding:"omitempty"`
	Append    *bool   `form:"append,default=0" binding:"omitempty"`
	Hidden    *bool   `form:"hidden,default=0" binding:"omitempty"`
}

// FileCreateHandler is used to create file or directory
func FileCreateHandler(ctx *gin.Context) {
	var (
		fh     *multipart.FileHeader
		err    error
		buf    = bytes.NewBuffer(nil)
		reader io.Reader

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}

		db            = ctx.MustGet("db").(*gorm.DB)
		ip            = ctx.ClientIP()
		input         = ctx.MustGet("inputParam").(*fileCreateInput)
		fileCreateSrv = &service.FileCreate{
			BaseService: service.BaseService{
				DB: db,
			},
			IP:    &ip,
			Path:  input.Path,
			Token: ctx.MustGet("token").(*models.Token),
		}

		fileCreateValue interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if fh, err = ctx.FormFile("file"); err != nil {
		if err != http.ErrMissingFile {
			reErrors = generateErrors(err, "file")
			return
		}
	} else {
		if reader, err = fh.Open(); err != nil {
			reErrors = generateErrors(err, "file")
			return
		}

		if _, err = io.Copy(buf, reader); err != nil {
			reErrors = generateErrors(err, "")
			return
		}

		if buf.Len() > models.ChunkSize {
			reErrors = generateErrors(models.ErrChunkExceedLimit, "file")
			return
		}

		// here, assign buf to reader, because the origin reader is multipart.sectionReadCloser
		// and the content of the origin reader has exhausted
		reader = buf

		if input.Hash != nil || input.Size != nil {

			if input.Size != nil && buf.Len() != *input.Size {
				reErrors = generateErrors(errors.New("the size of file doesn't match"), "size")
				return
			}
			if input.Hash != nil {
				var (
					h   string
					err error
				)
				if h, err = util.Sha256Hash2String(buf.Bytes()); err != nil {
					reErrors = generateErrors(err, "")
					return
				}
				if h != *input.Hash {
					reErrors = generateErrors(errors.New("the hash of file doesn't match"), "hash")
					return
				}
			}
		}
	}
	fileCreateSrv.Reader = reader
	if input.Hidden != nil && *input.Hidden {
		fileCreateSrv.Hidden = 1
	}
	if input.Overwrite != nil && *input.Overwrite {
		fileCreateSrv.Overwrite = 1
	}
	if input.Append != nil && *input.Append {
		fileCreateSrv.Append = 1
	}
	if input.Rename != nil && *input.Rename {
		fileCreateSrv.Rename = 1
	}

	if isTesting {
		fileCreateSrv.RootPath = testingChunkRootPath
	}

	if err := fileCreateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if fileCreateValue, err = fileCreateSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	if data, err = fileResp(fileCreateValue.(*models.File), db); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	code = 200
	success = true
}
