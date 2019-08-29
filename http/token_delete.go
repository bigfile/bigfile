//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"github.com/bigfile/bigfile/databases/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type tokenDeleteInput struct {
	AppUID string `form:"appUid" binding:"required"`
	Token  string `form:"token" binding:"required"`
	Nonce  string `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign   string `form:"sign" binding:"required"`
}

// TokenDeleteHandler is used to delete a token
func TokenDeleteHandler(ctx *gin.Context) {
	var (
		db    *gorm.DB
		err   error
		token *models.Token
		input *tokenDeleteInput

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

	input = ctx.MustGet("inputParam").(*tokenDeleteInput)
	db = ctx.MustGet("db").(*gorm.DB)

	if token, err = models.FindTokenByUID(input.Token, db); err != nil {
		reErrors = generateErrors(err, "token")
		return
	}
	db.Delete(token)
	db.Unscoped().First(token)
	success = true
	code = 200
	data = tokenResp(token)
}
