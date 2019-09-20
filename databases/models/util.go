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
)

// SecretLength set the secret length
const SecretLength = 12

// NewSecret is used generate secret for app and token
func NewSecret() string {
	r := "1234567890abcdefghijklmnopqrstuvwxyzQWERTYUIOPASDFGHJKLZXCVBNM"
	randomBytes := make([]byte, SecretLength)
	rLength := len(r)
	mrand.Seed(time.Now().Unix())
	for i := 0; i < SecretLength; i++ {
		randomBytes[i] = r[mrand.Intn(rLength)]
	}
	return string(randomBytes)
}

// Random is used to generate random bytes
func Random(length uint) []byte {
	var r = make([]byte, length)
	_, _ = rand.Reader.Read(r)
	return r
}

// RandomWithMD5 is used to generate random and hashed by md5
func RandomWithMD5(length uint) string {
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
