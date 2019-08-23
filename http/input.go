//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package http

// AppUIDInput represent 'AppUid' request param
type AppUIDInput struct {
	AppUID string `form:"appUid" binding:"required"`
}

// NonceInput represent `nonce` request param
type NonceInput struct {
	Nonce *string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
}

// TokenInput represent `token` request param
type TokenInput struct {
	Token string `form:"token" binding:"required"`
}
