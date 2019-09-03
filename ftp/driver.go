//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package ftp

import (
	"fmt"
	"strings"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/jinzhu/gorm"
	"goftp.io/server"
)

// Driver is used to operate files
type Driver struct {
	db       *gorm.DB
	app      *models.App
	conn     *server.Conn
	rootPath *string
	rootDir  *models.File
}

func (d *Driver) Init(conn *server.Conn) {
	d.conn = conn
}

// buildPath is used to build the real
func (d *Driver) buildPath(path string) string {
	if d.app == nil && d.rootPath == nil {
		loginUserName := d.conn.LoginUser()
		if strings.HasPrefix(loginUserName, tokenPrefix) {
			tokenUid := strings.TrimPrefix(loginUserName, tokenPrefix)
			token, _ := models.FindTokenByUID(tokenUid, d.db)
			d.app = &token.App
			d.rootPath = &token.Path
			d.rootDir, _ = models.CreateOrGetLastDirectory(d.app, token.Path, d.db)
		} else {
			d.app, _ = models.FindAppByUID(loginUserName, d.db)
			rootPath := "/"
			d.rootPath = &rootPath
			d.rootDir, _ = models.CreateOrGetRootPath(d.app, d.db)
		}
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(*d.rootPath, "/"), strings.TrimPrefix(path, "/"))
}
