//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

// test cors
func TestRouters(t *testing.T) {
	var (
		w   = httptest.NewRecorder()
		api = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create")
	)
	trx, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = trx
	config.DefaultConfig.CORSEnable = true
	req := httptest.NewRequest("OPTIONS", api, strings.NewReader(""))
	req.Header.Set("Origin", "192.168.0.1")
	Routers().ServeHTTP(w, req)
	headers := w.Header()
	assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
}

func TestRouters2(t *testing.T) {
	var (
		w     = httptest.NewRecorder()
		err   error
		trx   *gorm.DB
		api   = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/delete")
		token *models.Token
		down  func(*testing.T)
	)
	config.DefaultConfig.LimitRateByIPEnable = true

	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDBConn = trx

	params := map[string]interface{}{
		"appUid": token.App.UID,
		"token":  token.UID,
		"nonce":  models.RandomWithMd5(222),
	}

	apiWithQs := fmt.Sprintf("%s?%s", api, getParamsSignBody(params, token.App.Secret))
	req, _ := http.NewRequest("DELETE", apiWithQs, nil)
	Routers().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouters3(t *testing.T) {
	var (
		w   = httptest.NewRecorder()
		api = buildRoute(config.DefaultConfig.HTTP.APIPrefix, "/token/create")
	)
	trx, down := models.SetUpTestCaseWithTrx(nil, t)
	defer down(t)
	testDBConn = trx

	config.DefaultConfig.CORSEnable = true
	config.DefaultConfig.AccessLogFile = filepath.Join(os.TempDir(), fmt.Sprintf("%s.txt", models.RandomWithMd5(22)))
	config.DefaultConfig.Log.File.Enable = false

	req := httptest.NewRequest("OPTIONS", api, strings.NewReader(""))
	req.Header.Set("Origin", "192.168.0.1")
	Routers().ServeHTTP(w, req)
	headers := w.Header()
	assert.Equal(t, "*", headers.Get("Access-Control-Allow-Origin"))
	assert.True(t, util.IsFile(config.DefaultConfig.AccessLogFile))
	assert.Nil(t, os.Remove(config.DefaultConfig.AccessLogFile))
	config.DefaultConfig.AccessLogFile = ""
}
