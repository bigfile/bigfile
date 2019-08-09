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
	return func(c *gin.Context) {
		var (
			input struct {
				AppId string `form:"appId" json:"appId" binding:"required"`
			}
			app   *models.App
			err   error
			ok    bool
			ctxDb interface{}
		)
		if err = c.ShouldBind(&input); err == nil {
			if ctxDb, ok = c.Get("db"); ok {
				if app, err = models.FindAppByUID(input.AppId, ctxDb.(*gorm.DB)); err == nil {
					c.Set("app", app)
				} else {
					c.AbortWithStatusJSON(400, &Response{
						RequestId: c.GetInt64("requestId"),
						Success:   false,
						Errors: map[string][]string{
							"appId": {"cant't parse app from AppId"},
						},
						Data: nil,
					})
				}
			}
		} else {
			c.AbortWithStatusJSON(400, &Response{
				RequestId: c.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"appId": {err.Error()},
				},
				Data: nil,
			})
		}
		c.Next()
	}
}

// ConfigContextMiddleware will config context for each request. such as: db connection
func ConfigContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		if db == nil {
			db = databases.MustNewConnection(&config.DefaultConfig.Database)
		}
		context.Set("db", db)
		context.Next()
	}
}

// SignMiddlewareWithApp will validate request signature of request
func SignMiddlewareWithApp(input interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBind(input); err != nil {
			c.AbortWithStatusJSON(400, &Response{
				RequestId: c.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"inputParamError": {err.Error()},
				},
			})

		} else {
			c.Set("inputParam", input)
			appValue, _ := c.Get("app")
			app := appValue.(*models.App)
			if !ValidateRequestSignature(c, app.Secret) {
				c.AbortWithStatusJSON(400, &Response{
					RequestId: c.GetInt64("requestId"),
					Success:   false,
					Errors: map[string][]string{
						"sign": {"request param sign error"},
					},
				})
			}
		}
		c.Next()
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

	return hex.EncodeToString(m.Sum(nil)) == secret
}
