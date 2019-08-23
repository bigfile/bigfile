//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	pathToFileCache = cache.New(5*time.Minute, 10*time.Minute)
)
