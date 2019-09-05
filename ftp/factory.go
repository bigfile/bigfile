//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package ftp

import (
	"github.com/bigfile/bigfile/databases"
	"goftp.io/server"
)

// Factory is a driver factory, is used to generate driver when new connection comes
type Factory struct{}

// NewDriver return a driver
func (factory *Factory) NewDriver() (server.Driver, error) {
	return &Driver{db: databases.MustNewConnection(nil)}, nil
}
