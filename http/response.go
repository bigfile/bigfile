//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

import (
	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/bigfile/bigfile/service"
)

// Response represent http response for client
type Response struct {
	RequestID int64               `json:"requestId"`
	Success   bool                `json:"success"`
	Errors    map[string][]string `json:"errors"`
	Data      interface{}         `json:"data"`
}

// generateErrors is used to generate errors
func generateErrors(err error) map[string][]string {

	if err == nil {
		return nil
	}

	if vErr, ok := err.(service.ValidateErrors); ok {
		return vErr.MapFieldErrors()
	}
	return map[string][]string{
		"system": {err.Error()},
	}
}

// tokenResp is sed to generate token json response
func tokenResp(token *models.Token) map[string]interface{} {

	var expiredAt interface{} = token.ExpiredAt

	if token.ExpiredAt != nil {
		expiredAt = token.ExpiredAt.Unix()
	}

	return map[string]interface{}{
		"token":          token.UID,
		"ip":             token.IP,
		"availableTimes": token.AvailableTimes,
		"readOnly":       token.ReadOnly,
		"expiredAt":      expiredAt,
		"path":           token.Path,
		"secret":         token.Secret,
	}
}
