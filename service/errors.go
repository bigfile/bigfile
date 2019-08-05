//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package service

import "fmt"

// ValidateError is defined validate error information
type ValidateError struct {
	Msg   string
	Field string
	Code  int
	Desc  string
}

// Error implement error interface
func (v *ValidateError) Error() string {
	return fmt.Sprintf("%s, validate error: %s %s", v.Field, v.Msg, v.Desc)
}

var (
	// ErrMap map service field to specific error
	ErrMap = map[string]*ValidateError{

		"Unknown": {
			Code: 10000, Field: "Unknown", Msg: "unknown error",
		},

		"App": {
			Code: 10001, Field: "App", Msg: "can't find specific application by input params",
		},

		"TokenCreate.App": {
			Code: 10002, Field: "Token.App", Msg: "can't find specific application by input params",
		},
		"TokenCreate.Path": {
			Code: 10003, Field: "Token.Path", Msg: "path of token can't be empty",
		},
		"TokenCreate.Ip": {
			Code: 10003, Field: "Token.Ip", Msg: "max length of ip is 1500",
		},
		"TokenCreate.Secret": {
			Code: 10004, Field: "Token.Secret", Msg: "secret of token is 32",
		},
		"TokenCreate.AvailableTimes": {
			Code: 10005, Field: "Token.AvailableTimes", Msg: "availableTimes of token is greater than -1",
		},
		"TokenCreate.ReadOnly": {
			Code: 10006, Field: "Token.ReadOnly", Msg: "readOnly of token is 0 or 1",
		},
	}

	// ErrApp is equal to ErrMap["App"]
	ErrApp = ErrMap["App"]

	// ErrUnknown is equal to ErrMap["UnknownError"]
	ErrUnknown = ErrMap["UnknownError"]
)
