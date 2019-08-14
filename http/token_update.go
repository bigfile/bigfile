//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"context"
	"reflect"
	"time"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type tokenUpdateInput struct {
	AppUID         string     `form:"appUid" binding:"required"`
	Token          string     `form:"token" binding:"required"`
	Nonce          string     `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign           string     `form:"sign" binding:"required"`
	Path           *string    `form:"path" binding:"omitempty,max=1000"`
	IP             *string    `form:"ip" binding:"omitempty,max=1500"`
	ExpiredAt      *time.Time `form:"expiredAt" time_format:"unix" binding:"omitempty,gt"`
	Secret         *string    `form:"secret" binding:"omitempty,len=32"`
	AvailableTimes *int       `form:"availableTimes" binding:"omitempty,max=2147483647"`
	ReadOnly       *bool      `form:"readOnly"`
}

// TokenUpdateHandler is used to handle request for update token
func TokenUpdateHandler(ctx *gin.Context) {
	var (
		input            *tokenUpdateInput
		db               *gorm.DB
		tokenUpdateSrv   *service.TokenUpdate
		readOnlyI8       int8
		err              error
		tokenUpdateValue interface{}

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

	input = ctx.MustGet("inputParam").(*tokenUpdateInput)
	db = ctx.MustGet("db").(*gorm.DB)

	if input.ReadOnly != nil && *input.ReadOnly {
		readOnlyI8 = 1
	}

	tokenUpdateSrv = &service.TokenUpdate{
		BaseService: service.BaseService{
			DB: db,
		},
		Token:          input.Token,
		Secret:         input.Secret,
		Path:           input.Path,
		IP:             input.IP,
		ExpiredAt:      input.ExpiredAt,
		AvailableTimes: input.AvailableTimes,
		ReadOnly:       &readOnlyI8,
	}

	if err = tokenUpdateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err)
		return
	}

	if tokenUpdateValue, err = tokenUpdateSrv.Execute(context.TODO()); err != nil {
		reErrors = generateErrors(err)
		return
	}

	code = 200
	success = true
	data = tokenResp(tokenUpdateValue.(*models.Token))
}
