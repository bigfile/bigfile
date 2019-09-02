//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateErrors_ContainsErrCode(t *testing.T) {
	var (
		err1   = &ValidateError{Code: 11111}
		err2   = &ValidateError{Code: 22222}
		errors = ValidateErrors{err1, err2}
	)
	assert.True(t, errors.ContainsErrCode(err1.Code))
	assert.True(t, errors.ContainsErrCode(err2.Code))
	assert.False(t, errors.ContainsErrCode(33333))
}

func TestValidateErrors_Error(t *testing.T) {
	var (
		err1   = &ValidateError{Code: 11111}
		err2   = &ValidateError{Code: 22222}
		errors = ValidateErrors{err1, err2}
		ok     bool
	)
	assert.Contains(t, errors.Error(), "11111")
	assert.Contains(t, errors.Error(), "22222")
	_, ok = interface{}(errors).(error)
	assert.True(t, ok)

}

func TestValidateErrors_Map(t *testing.T) {
	var (
		err1   = &ValidateError{Code: 11111}
		err2   = &ValidateError{Code: 22222}
		errors = ValidateErrors{err1, err2}
		m      = errors.Map()
		ok     bool
	)
	_, ok = m[err1.Code]
	assert.True(t, ok)
	_, ok = m[err2.Code]
	assert.True(t, ok)
	_, ok = m[333333333]
	assert.False(t, ok)
}

func TestValidateError_Error(t *testing.T) {
	var err1 = &ValidateError{Code: 11111}
	assert.Contains(t, err1.Error(), "11111")
	_, ok := interface{}(err1).(error)
	assert.True(t, ok)
}

func TestValidateErrors_MapFieldErrors(t *testing.T) {
	var err = ValidateErrors{&ValidateError{
		Msg:   "1111",
		Field: "field",
		Code:  111,
	}}
	assert.Equal(t, "code: 111, field: field, validate error: 1111", err.MapFieldErrors()["field"][0])
}
