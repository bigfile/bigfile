//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bigfile/bigfile/config"
)

func TestTokenCreateHandler(t *testing.T) {
	var (
		buf           = new(bytes.Buffer)
		multipartBody = multipart.NewWriter(buf)
	)

	_ = multipartBody.WriteField("appId", "10001")
	_ = multipartBody.WriteField("requestTime", strconv.FormatInt(time.Now().Unix(), 10))
	_ = multipartBody.WriteField("sign", "sign")
	_ = multipartBody.WriteField("secret", "secret111")
	multipartBody.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create"), buf)
	req.Header.Set("Content-Type", multipartBody.FormDataContentType())
	Routers().ServeHTTP(w, req)
}
