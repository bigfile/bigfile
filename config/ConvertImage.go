//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package config

// ConvertImage represent config for convertImage
type ConvertImage struct {
	Cache     bool   `yaml:"cache,omitempty"`
	CachePath string `yaml:"cachePath,omitempty"`
}
