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
	"github.com/bigfile/bigfile/databases/models"
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

// RecordRequestMiddleware is used to record request in database.
// It's should be put behind ConfigContextMiddleware
func RecordRequestMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			db        = ctx.MustGet("db").(*gorm.DB)
			bw        = &bodyWriter{ResponseWriter: ctx.Writer, body: bytes.NewBufferString("")}
			reqRecord = models.MustNewHTTPRequest(ctx.ClientIP(), ctx.Request.Method, ctx.Request.URL.String(), db)
		)

		ctx.Writer = bw
		ctx.Set("requestId", int64(reqRecord.ID))
		ctx.Set("reqRecord", reqRecord)
		ctx.Next()
		// If some handle return file stream data, that should not
		// written to database.
		if _, ok := ctx.Get("ignoreRespBody"); !ok {
			reqRecord.ResponseBody = bw.body.String()
		} else {
			bw.body.Reset()
		}
		reqBodyString, _ := janitor.MarshalToString(ctx.Request.Form)
		reqRecord.RequestBody = reqBodyString
		reqRecord.ResponseCode = ctx.Writer.Status()
		reqHeaderString, _ := janitor.MarshalToString(ctx.Request.Header)
		reqRecord.RequestHeader = reqHeaderString
		_ = reqRecord.Save(db)
	}
}

// ParseAppMiddleware will parse request context to get an app.
// It's should be put behind RecordRequestMiddleware
func ParseAppMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			input AppUIDInput
			app   *models.App
			err   error
			ok    bool
			ctxDb interface{}
		)
		if err = ctx.ShouldBind(&input); err == nil {
			if ctxDb, ok = ctx.Get("db"); ok {
				if app, err = models.FindAppByUID(input.AppUID, ctxDb.(*gorm.DB)); err == nil {
					reqRecord := ctx.MustGet("reqRecord").(*models.Request)
					reqRecord.AppID = &app.ID
					ctx.Set("app", app)
				} else {
					ctx.AbortWithStatusJSON(400, &Response{
						RequestID: ctx.GetInt64("requestId"),
						Success:   false,
						Errors: map[string][]string{
							"appUid": {"cant't parse app from appUid"},
						},
					})
				}
			}
		} else {
			ctx.AbortWithStatusJSON(400, &Response{
				RequestID: ctx.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"appUid": {err.Error()},
				},
			})
		}
		ctx.Next()
	}
}

// ParseTokenMiddleware is used to parse request context and get a token
// It's should be put behind RecordRequestMiddleware
func ParseTokenMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			db        = ctx.MustGet("db").(*gorm.DB)
			err       error
			input     TokenInput
			token     *models.Token
			requestID = ctx.GetInt64("requestId")
			reqRecord = ctx.MustGet("reqRecord").(*models.Request)
		)
		if err = ctx.ShouldBind(&input); err == nil {
			if token, err = models.FindTokenByUID(input.Token, db); err != nil {
				ctx.AbortWithStatusJSON(400, &Response{
					RequestID: requestID,
					Success:   false,
					Errors: map[string][]string{
						"token": {"token find failed"},
					},
				})
			} else {
				reqRecord.Token = &token.UID
				reqRecord.AppID = &token.AppID
				ctx.Set("app", &token.App)
				ctx.Set("token", token)
			}
		} else {
			ctx.AbortWithStatusJSON(400, &Response{
				RequestID: requestID,
				Success:   false,
				Errors: map[string][]string{
					"token": {err.Error()},
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

// SignWithAppMiddleware will validate request signature of request.
// It's should be put behind ParseAppMiddleware
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
			app := ctx.MustGet("app").(*models.App)
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

// SignWithTokenMiddleware will validate request signature of request.
// It's should be put behind SignWithTokenMiddleware
func SignWithTokenMiddleware(input interface{}) gin.HandlerFunc {
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
			token := ctx.MustGet("token").(*models.Token)
			if token.Secret != nil && !ValidateRequestSignature(ctx, *token.Secret) {
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

// RateLimitByIPMiddleware will record requests number by ip, avoid floor attack.
// It's should put behind RecordRequestMiddleware
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

// ReplayAttackMiddleware is used to avoid request replay attack.
// We recommend that you should provide a 'nonce' value for request
// in all of 'UPDATE' request. But for 'QUERY' request, you should
// not add this. Of course, if you want to do this, bigfile allow that.
//
// For developers, this middleware should be put behind ParseAppMiddleware
// or ParseTokenMiddleware, we need appId to validate this.
func ReplayAttackMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			db        = ctx.MustGet("db").(*gorm.DB)
			app       = ctx.MustGet("app").(*models.App)
			reqRecord = ctx.MustGet("reqRecord").(*models.Request)
			input     NonceInput
			err       error
		)
		if err = ctx.ShouldBind(&input); err == nil {
			if input.Nonce != nil {
				if t, err := models.FindRequestWithAppAndNonce(app, *input.Nonce, db); err == nil && t.ID > 0 {
					ctx.AbortWithStatusJSON(400, &Response{
						RequestID: ctx.GetInt64("requestId"),
						Success:   false,
						Errors: map[string][]string{
							"nonce": {"this request is being replayed"},
						},
					})
				}
				reqRecord.Nonce = input.Nonce
			}
		} else {
			ctx.AbortWithStatusJSON(400, &Response{
				RequestID: ctx.GetInt64("requestId"),
				Success:   false,
				Errors: map[string][]string{
					"nonce": {"nonce is optional, but the min length of nonce is 32, the max length is 48"},
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
