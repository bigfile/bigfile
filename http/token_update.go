//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"reflect"
	"time"

	"github.com/bigfile/bigfile/service"

	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
)

type tokenUpdateInput struct {
	AppUID         string     `form:"appUid" json:"appUid" binding:"required"`
	Token          string     `form:"token" json:"token" binding:"required"`
	Nonce          string     `form:"nonce" json:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign           string     `form:"sign" json:"sign" binding:"required"`
	Path           *string    `form:"path,default=/" json:"path,default=/" binding:"max=1000"`
	IP             *string    `form:"ip" json:"ip" binding:"omitempty,max=1500"`
	ExpiredAt      *time.Time `form:"expiredAt" json:"expiredAt" time_format:"unix" binding:"omitempty,gt"`
	Secret         *string    `form:"secret" json:"secret" binding:"omitempty,len=32"`
	AvailableTimes *int       `form:"availableTimes,default=-1" json:"availableTimes,default=-1" binding:"omitempty,max=2147483647"`
	ReadOnly       *bool      `form:"readOnly,default=0" json:"readOnly,default=0"`
}

// TokenUpdateHandler is used to handle request for update token
func TokenUpdateHandler(ctx *gin.Context) {
	var (
		input          *tokenUpdateInput
		db             *gorm.DB
		tokenUpdateSrv *service.TokenUpdate
		readOnlyI8     int8
		err            error
	)

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
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors: map[string][]string{
				"system": {err.Error()},
			},
		})
		return

	}

}
