//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Request is used to record client request
type Request struct {
	ID           uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Protocol     string    `gorm:"type:CHAR(10) NOT NULL;column:protocol"`
	AppID        *uint64   `gorm:"type:BIGINT(20) UNSIGNED;DEFAULT:NULL;column:appId"`
	Token        *string   `gorm:"type:CHAR(32);DEFAULT:NULL;column:token"`
	IP           *string   `gorm:"type:CHAR(15);column:ip;DEFAULT:NULL"`
	Method       *string   `gorm:"type:CHAR(10);column:method;DEFAULT:NULL"`
	Service      *string   `gorm:"type:VARCHAR(512);column:service;DEFAULT:NULL"`
	RequestBody  string    `gorm:"type:TEXT;column:requestBody"`
	ResponseCode int       `gorm:"type:int;column:responseCode;DEFAULT:200"`
	ResponseBody string    `gorm:"type:TEXT;column:responseBody"`
	CreatedAt    time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
}

// Save is used to persistent update to database
func (r *Request) Save(db *gorm.DB) error {
	return db.Save(r).Error
}

// NewRequestWithProtocol is used to generate new request record
func NewRequestWithProtocol(protocol string, db *gorm.DB) (*Request, error) {
	var (
		req = &Request{
			Protocol: protocol,
		}
		err error
	)
	err = db.Create(req).Error
	return req, err
}

// MustNewRequestWithProtocol is used to generate new request record. But, if
// some errors happened, it will panic
func MustNewRequestWithProtocol(protocol string, db *gorm.DB) *Request {
	if req, err := NewRequestWithProtocol(protocol, db); err != nil {
		panic(err)
	} else {
		return req
	}
}

// MustNewRequestWithHTTPProtocol is used to generate http protocol request record.
func MustNewRequestWithHTTPProtocol(ip, method, url string, db *gorm.DB) *Request {
	var (
		req = &Request{
			Protocol: "http",
			IP:       &ip,
			Method:   &method,
			Service:  &url,
		}
		err error
	)
	err = db.Create(req).Error
	if err != nil {
		panic(err)
	}
	return req
}
