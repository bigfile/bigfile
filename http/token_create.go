//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"context"
	"time"

	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/bigfile/bigfile/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type tokenCreateInput struct {
	AppUID         string     `form:"appUid" json:"appUid" binding:"required"`
	RequestTime    time.Time  `form:"requestTime" json:"requestTime" time_format:"unix" binding:"required"`
	Sign           string     `form:"sign" json:"sign" binding:"required"`
	Path           *string    `form:"path,default=/" json:"path,default=/" binding:"max=1000"`
	IP             *string    `form:"ip" json:"ip" binding:"omitempty,max=1500"`
	ExpiredAt      *time.Time `form:"expiredAt" json:"expiredAt" time_format:"unix" binding:"omitempty,gt"`
	Secret         *string    `form:"secret" json:"secret" binding:"omitempty,len=32"`
	AvailableTimes *int       `form:"availableTimes,default=-1" json:"availableTimes,default=-1" binding:"omitempty,max=2147483647"`
	ReadOnly       *bool      `form:"readOnly,default=0" json:"readOnly,default=0"`
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
	)
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

	if err := tokenCreateSrv.Validate(); err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors:    err.MapFieldErrors(),
		})
		return
	}

	if tokenCreateValue, err = tokenCreateSrv.Execute(context.Background()); err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors: map[string][]string{
				"system": {err.Error()},
			},
		})
		return
	}

	ctx.JSON(200, &Response{
		RequestID: ctx.GetInt64("requestId"),
		Success:   true,
		Data:      tokenCreateValue.(*models.Token),
	})
}
