//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenCreateInput struct {
	AppId          uint64     `form:"appId" json:"appId" binding:"required"`
	RequestTime    time.Time  `form:"requestTime" json:"requestTime" time_format:"unix" binding:"required"`
	Sign           string     `form:"sign" json:"sign" binding:"required"`
	Path           *string    `form:"path" json:"path"`
	IP             *string    `form:"ip" json:"ip"`
	ExpiredAt      *time.Time `form:"expiredAt" json:"expiredAt"`
	Secret         *string    `form:"secret" json:"secret"`
	AvailableTimes *int8      `form:"availableTimes" json:"availableTimes"`
	ReadOnly       *bool      `form:"readOnly" json:"readOnly"`
}

func TokenCreateHandler(ctx *gin.Context) {
	var input tokenCreateInput
	if err := ctx.ShouldBind(&input); err != nil {
		fmt.Println(err)
	}
	fmt.Println(*input.Secret)
}
