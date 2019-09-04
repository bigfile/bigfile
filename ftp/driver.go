//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package ftp

import (
	"fmt"
	"io"
	"strings"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/jinzhu/gorm"
	"goftp.io/server"
)

var appRootPath = "/"

// Driver is used to operate files
type Driver struct {
	db            *gorm.DB
	app           *models.App
	conn          *server.Conn
	rootPath      *string
	rootDir       *models.File
	rootChunkPath *string
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
			d.rootPath = &appRootPath
			d.rootDir, _ = models.CreateOrGetRootPath(d.app, d.db)
		}
	}
	return strings.TrimSuffix(
		fmt.Sprintf(
			"%s/%s",
			strings.TrimSuffix(*d.rootPath, "/"),
			strings.TrimPrefix(path, "/"),
		),
		"/",
	)
}

// Stat will return the information by the path
func (d *Driver) Stat(path string) (fileInfo server.FileInfo, err error) {
	var file *models.File
	if file, err = models.FindFileByPath(d.app, d.buildPath(path), d.db, true); err != nil {
		return
	}
	return &FileInfo{
		name:     file.Name,
		size:     int64(file.Size),
		isDir:    file.IsDir == models.IsDir,
		modeTime: file.UpdatedAt,
	}, nil
}

// ChangeDir is used to toggle current directory, if the directory doesn't exist,
// it will be created.
func (d *Driver) ChangeDir(path string) (err error) {
	_, err = models.CreateOrGetLastDirectory(d.app, d.buildPath(path), d.db)
	return err
}

// ListDir is used to list files and subDir of current dir
func (d *Driver) ListDir(path string, callback func(server.FileInfo) error) (err error) {
	var dir *models.File
	if dir, err = models.CreateOrGetLastDirectory(d.app, d.buildPath(path), d.db); err != nil {
		return
	}
	if err = d.db.Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Order("isDir DESC")
	}).First(dir).Error; err != nil {
		return
	}
	for _, child := range dir.Children {
		if err = callback(&FileInfo{
			name: child.Name, size: int64(child.Size),
			isDir: child.IsDir == models.IsDir, modeTime: child.UpdatedAt}); err != nil {
			return
		}
	}
	return
}

// DeleteDir is used to delete a directory
func (d *Driver) DeleteDir(path string) (err error) {
	var dir *models.File
	if dir, err = models.FindFileByPath(d.app, d.buildPath(path), d.db, true); err != nil {
		return
	}
	return dir.Delete(true, d.db)
}

// DeleteFile
func (d *Driver) DeleteFile(path string) (err error) {
	var file *models.File
	if file, err = models.FindFileByPath(d.app, d.buildPath(path), d.db, true); err != nil {
		return
	}
	return file.Delete(true, d.db)
}

// Rename is used to move file or rename file
func (d *Driver) Rename(fromPath string, toPath string) (err error) {
	var file *models.File
	if file, err = models.FindFileByPath(d.app, d.buildPath(fromPath), d.db, true); err != nil {
		return
	}
	return file.MoveTo(d.buildPath(toPath), d.db)
}

// MakeDir is used to create dir
func (d *Driver) MakeDir(path string) (err error) {
	_, err = models.CreateOrGetLastDirectory(d.app, d.buildPath(path), d.db)
	return err
}

// PutFile is used to upload file
func (d *Driver) PutFile(path string, dataConn io.Reader, append bool) (bytes int64, err error) {
	var file *models.File
	if append {
		if file, err = models.FindFileByPath(d.app, d.buildPath(path), d.db, true); err != nil {
			return
		}
		originSize := file.Size
		if err = file.AppendFromReader(dataConn, 0, d.rootChunkPath, d.db); err != nil {
			return
		}
		return int64(file.Size - originSize), nil
	} else {
		if file, err = models.CreateFileFromReader(
			d.app, d.buildPath(path), dataConn, 0, d.rootChunkPath, d.db); err != nil {
			return
		}
		return int64(file.Size), nil
	}
}
