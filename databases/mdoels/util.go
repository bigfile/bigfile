//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/log"
)

// NewSecret is used generate secret for app and token
func NewSecret() string {
	var (
		b           = make([]byte, 128)
		hash        = md5.New()
		letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)
	if _, err := rand.Reader.Read(b); err != nil {
		log.MustNewLogger(&config.DefaultConfig.Log).Warning(err)
		for i := range b {
			b[i] = letterBytes[mrand.Intn(len(letterBytes))]
		}
	}
	_, _ = hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}
