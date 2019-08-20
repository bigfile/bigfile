//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"reflect"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type tokenCreateInput struct {
	AppUID         string     `form:"appUid" binding:"required"`
	Nonce          string     `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign           string     `form:"sign" binding:"required"`
	Path           *string    `form:"path,default=/" binding:"max=1000"`
	IP             *string    `form:"ip" binding:"omitempty,max=1500"`
	ExpiredAt      *time.Time `form:"expiredAt" time_format:"unix" binding:"omitempty,gt"`
	Secret         *string    `form:"secret" binding:"omitempty,len=32"`
	AvailableTimes *int       `form:"availableTimes,default=-1" binding:"omitempty,max=2147483647"`
	ReadOnly       *bool      `form:"readOnly,default=0"`
}

// TokenCreateHandler is used to handle token create http request
func TokenCreateHandler(ctx *gin.Context) {
	var (
		input            *tokenCreateInput
		db               *gorm.DB
		app              *models.App
		tokenCreateSrv   *service.TokenCreate
		readOnlyI8       int8
		tokenCreateValue interface{}
		err              error

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

	input = ctx.MustGet("inputParam").(*tokenCreateInput)
	db = ctx.MustGet("db").(*gorm.DB)
	app = ctx.MustGet("app").(*models.App)

	if input.ReadOnly != nil && *input.ReadOnly {
		readOnlyI8 = 1
	}

	tokenCreateSrv = &service.TokenCreate{
		BaseService: service.BaseService{
			DB: db,
		},
		IP:             input.IP,
		App:            app,
		Path:           *input.Path,
		Secret:         input.Secret,
		ReadOnly:       readOnlyI8,
		ExpiredAt:      input.ExpiredAt,
		AvailableTimes: *input.AvailableTimes,
	}

	if err := tokenCreateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err)
		return
	}

	if tokenCreateValue, err = tokenCreateSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err)
		return
	}

	data = tokenResp(tokenCreateValue.(*models.Token))
	success = true
	code = 200
}
