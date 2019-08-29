//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"
	"strconv"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/log"
)

// NewSecret is used generate secret for app and token
func NewSecret() string {
	return RandomWithMd5(32)
}

// Random is used to generate random bytes
func Random(length uint) []byte {
	var (
		r           = make([]byte, length)
		letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)
	if _, err := rand.Reader.Read(r); err != nil {
		log.MustNewLogger(&config.DefaultConfig.Log).Warning(err)
		for i := range r {
			r[i] = letterBytes[mrand.Intn(len(letterBytes))]
		}
	}
	return r
}

// RandomWithMd5 is used to generate random and hashed by md5
func RandomWithMd5(length uint) string {
	var (
		b    = Random(length)
		hash = md5.New()
	)
	_, _ = hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

// UID is used to generate uid
func UID() string {
	random := Random(32)
	random = append(random, []byte(strconv.FormatInt(time.Now().UnixNano(), 10))...)
	hash := md5.New()
	_, _ = hash.Write(random)
	return hex.EncodeToString(hash.Sum(nil))
}
