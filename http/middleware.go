//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	janitor "github.com/json-iterator/go"
	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

var (
	isTesting  bool
	testDBConn *gorm.DB
	limiterSet = cache.New(5*time.Minute, 10*time.Minute)
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
		db := ctx.MustGet("db").(*gorm.DB)
		reqRecord := models.MustNewRequestWithHTTPProtocol(
			ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL.String(), db,
		)
		ctx.Set("requestId", int64(reqRecord.ID))
		ctx.Set("reqRecord", reqRecord)
		ctx.Next()
		// If some handle return file stream data, that should not
		// written to database.
		if _, ok := ctx.Get("ignoreRespBody"); !ok {
			reqBodyString, _ := janitor.MarshalToString(ctx.Request.Form)
			reqRecord.RequestBody = reqBodyString
			reqRecord.ResponseCode = ctx.Writer.Status()
			reqRecord.ResponseBody = bw.body.String()
			_ = reqRecord.Save(db)
		}
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
					reqRecord := ctx.MustGet("reqRecord").(*models.Request)
					reqRecord.AppID = &app.ID
					ctx.Set("app", app)
				} else {
					ctx.AbortWithStatusJSON(400, &Response{
						RequestID: ctx.GetInt64("requestId"),
						Success:   false,
						Errors: map[string][]string{
							"appId": {"cant't parse app from AppId"},
						},
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
			})
		}
		ctx.Next()
	}
}

// ConfigContextMiddleware will config context for each request. such as: db connection
func ConfigContextMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if db == nil {
			if isTesting {
				db = testDBConn
			} else {
				db = databases.MustNewConnection(&config.DefaultConfig.Database)
			}
		}
		ctx.Set("db", db)
		ctx.Next()
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

// AccessLogMiddleware just wrap gin.LoggerWithFormatter
func AccessLogMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RateLimitByIPMiddleware will requests number by ip, avoid being attacked
func RateLimitByIPMiddleware(interval time.Duration, maxNumber int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		limiter, ok := limiterSet.Get(ip)
		if !ok {
			var expire = interval * 10
			limiter = rate.NewLimiter(rate.Every(interval), maxNumber)
			limiterSet.Set(ip, limiter, expire)
		}
		if !limiter.(*rate.Limiter).Allow() {
			ctx.AbortWithStatusJSON(429, &Response{
				RequestID: ctx.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"limitRateByIp": {"too many requests"},
				},
			})
		}
		ctx.Next()
	}
}

// SignStrWithSecret will calculate the sign of request paramStr that
// has already been sorted and concat together.
func SignStrWithSecret(paramStr, secret string) string {
	m := md5.New()
	_, _ = m.Write([]byte(paramStr))
	_, _ = m.Write([]byte(secret))
	return hex.EncodeToString(m.Sum(nil))
}
