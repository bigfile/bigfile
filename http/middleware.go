//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var isTesting bool

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (bw *bodyWriter) Write(p []byte) (int, error) {
	bw.body.Write(p)
	return bw.ResponseWriter.Write(p)
}

// RecordRequestMiddleware is used to record request in database
func RecordRequestMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bw := &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
		ctx.Writer = bw
		ctx.Set("requestId", int64(1000001))
		ctx.Next()
		// TODO: record response body
	}
}

// ParseAppMiddleware will parse request context to get an app
func ParseAppMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			input struct {
				AppID string `form:"appId" json:"appId" binding:"required"`
			}
			app   *models.App
			err   error
			ok    bool
			ctxDb interface{}
		)
		if err = ctx.ShouldBind(&input); err == nil {
			if ctxDb, ok = ctx.Get("db"); ok {
				if app, err = models.FindAppByUID(input.AppID, ctxDb.(*gorm.DB)); err == nil {
					ctx.Set("app", app)
				} else {
					ctx.AbortWithStatusJSON(400, &Response{
						RequestID: ctx.GetInt64("requestId"),
						Success:   false,
						Errors: map[string][]string{
							"appId": {"cant't parse app from AppId"},
						},
						Data: nil,
					})
				}
			}
		} else {
			ctx.AbortWithStatusJSON(400, &Response{
				RequestID: ctx.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"appId": {err.Error()},
				},
				Data: nil,
			})
		}
		ctx.Next()
	}
}

// ConfigContextMiddleware will config context for each request. such as: db connection
func ConfigContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var inTrx bool
		if db == nil {
			db = databases.MustNewConnection(&config.DefaultConfig.Database)
			if isTesting {
				db = db.Begin()
				inTrx = true
			}
		}
		ctx.Set("db", db)
		ctx.Next()
		if isTesting && inTrx {
			db.Rollback()
		}
	}
}

// SignWithAppMiddleware will validate request signature of request
func SignWithAppMiddleware(input interface{}) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := ctx.ShouldBind(input); err != nil {
			ctx.AbortWithStatusJSON(400, &Response{
				RequestID: ctx.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"inputParamError": {err.Error()},
				},
			})

		} else {
			ctx.Set("inputParam", input)
			appValue, _ := ctx.Get("app")
			app := appValue.(*models.App)
			if !ValidateRequestSignature(ctx, app.Secret) {
				ctx.AbortWithStatusJSON(400, &Response{
					RequestID: ctx.GetInt64("requestId"),
					Success:   false,
					Errors: map[string][]string{
						"sign": {"request param sign error"},
					},
				})
			}
		}
		ctx.Next()
	}
}

// ValidateRequestSignature will validate the signature of request
func ValidateRequestSignature(ctx *gin.Context, secret string) bool {

	var (
		params    = make(map[string]string)
		sign      = ctx.Request.FormValue("sign")
		keys      = make([]string, 1)
		signature = bytes.NewBufferString("")
		m         = md5.New()
	)

	if sign == "" {
		return false
	}

	for k, v := range ctx.Request.Form {
		if k != "sign" {
			params[k] = v[0]
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for index, k := range keys {
		if k != "" {
			signature.WriteString(k)
			signature.WriteString("=")
			signature.WriteString(params[k])
			if index != len(keys)-1 {
				signature.WriteString("&")
			}
		}
	}
	signature.WriteString(secret)
	_, _ = m.Write(signature.Bytes())

	return hex.EncodeToString(m.Sum(nil)) == sign
}

// SignStrWithSecret will calculate the sign of request paramStr that
// has already been sorted and concat together.
func SignStrWithSecret(paramStr, secret string) string {
	m := md5.New()
	_, _ = m.Write([]byte(paramStr))
	_, _ = m.Write([]byte(secret))
	return hex.EncodeToString(m.Sum(nil))
}
