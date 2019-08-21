//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
	"labix.org/v2/mgo/bson"
)

var (
	ErrFileExisted  = errors.New("file has already existed")
	ErrOverwriteDir = errors.New("directory can't be overwritten")
	ErrAppendToDir  = errors.New("can't append data to directory")
)

// File represent a file or a directory of system. If it's a file
// it has to associate with an object. Actually, the object hold
// the real content of file.
type File struct {
	ID            uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID           string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	PID           uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:pid"`
	AppID         uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:appId"`
	ObjectID      uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:objectId"`
	Size          int        `gorm:"type:int;column:size"`
	Name          string     `gorm:"type:VARCHAR(255);NOT NULL;column:name"`
	Ext           string     `gorm:"type:VARCHAR(255);NOT NULL;column:ext"`
	IsDir         int8       `gorm:"type:tinyint;column:isDir;DEFAULT:0"`
	Hidden        int8       `gorm:"type:tinyint;column:hidden;DEFAULT:0"`
	DownloadCount uint64     `gorm:"type:BIGINT(20);column:downloadCount;DEFAULT:0"`
	CreatedAt     time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt     time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt     *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`

	Object    Object    `gorm:"foreignkey:objectId"`
	App       App       `gorm:"foreignkey:appId"`
	Parent    *File     `gorm:"foreignkey:id;association_foreignkey:pid"`
	Histories []History `gorm:"foreignkey:fileId"`
}

// TableName represent the name of files table
func (f *File) TableName() string {
	return "files"
}

// Path is used to get the complete path of file
func (f *File) Path(db *gorm.DB) (string, error) {
	var (
		err     error
		parts   []string
		current = *f
	)
	for {
		parts = append(parts, current.Name)
		if current.PID == 0 {
			break
		}
		temp := &File{}
		if err = db.Where("id = ?", current.PID).Find(temp).Error; err != nil {
			return "", err
		}
		current = *temp
	}

	util.ReverseSlice(parts)

	return strings.Join(parts, "/"), nil
}

// UpdateParentSize is used to update parent size. note, size may be a negative number.
func (f *File) UpdateParentSize(size int, db *gorm.DB) error {
	if err := db.Model(f).UpdateColumn("size", gorm.Expr("size + ?", size)).Error; err != nil {
		return err
	}
	if f.PID == 0 {
		return nil
	}
	if err := db.Preload("Parent").Find(f).Error; err != nil {
		return err
	}
	return f.Parent.UpdateParentSize(size, db)
}

// OverWriteFromReader is used to overwrite the object
func (f *File) OverWriteFromReader(reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) error {

	if f.IsDir == 0 {
		return ErrOverwriteDir
	}

	var (
		err      error
		path     string
		object   *Object
		sizeDiff int
		history  = &History{
			ObjectID: f.ObjectID,
			FileID:   f.ID,
			Path:     path,
		}
	)

	if path, err = f.Path(db); err != nil {
		return err
	}
	history.Path = path

	if err := db.Save(history).Error; err != nil {
		return err
	}

	if object, err = CreateObjectFromReader(reader, rootPath, db); err != nil {
		return err
	}

	f.Object = *object
	f.ObjectID = object.ID
	f.Hidden = hidden
	sizeDiff = object.Size - f.Size
	f.Size += sizeDiff

	if err = db.Save(f).Error; err != nil {
		return err
	}

	if err = db.Preload("Parent").Find(f).Error; err != nil {
		return err
	}

	return f.Parent.UpdateParentSize(sizeDiff, db)
}

// AppendFromReader is used to append content from reader to file
func (f *File) AppendFromReader(reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) error {

	if f.IsDir == 1 {
		return ErrAppendToDir
	}

	var (
		err    error
		size   int
		object *Object
	)
	if err = db.Preload("Object").Preload("Parent").Find(f).Error; err != nil {
		return err
	}

	if object, size, err = f.Object.AppendFromReader(reader, rootPath, db); err != nil {
		return err
	}

	f.Hidden = hidden
	f.Size += size
	f.Object = *object
	f.ObjectID = object.ID

	if err = db.Save(f).Error; err != nil {
		return err
	}

	return f.Parent.UpdateParentSize(size, db)
}

// CreateOrGetLastDirectory is used to get last level directory
func CreateOrGetLastDirectory(app *App, path string, db *gorm.DB) (*File, error) {
	var (
		parent *File
		err    error
		parts  = strings.Split(strings.Trim(strings.TrimSpace(path), "/"), string(os.PathSeparator))
	)

	if parent, err = CreateOrGetRootPath(app, db); err != nil {
		return nil, err
	}

	for _, part := range parts {
		file := &File{}
		err = db.Where(
			&File{AppID: app.ID, PID: parent.ID, Name: part}).Assign(
			&File{IsDir: 1, UID: bson.NewObjectId().Hex()}).FirstOrCreate(file).Error
		if err != nil {
			return nil, err
		}
		parent = file
	}

	return parent, nil
}

// CreateOrGetRootPath is used to create or get root directory
func CreateOrGetRootPath(app *App, db *gorm.DB) (*File, error) {
	var (
		file = &File{}
		err  error
	)
	err = db.Where(
		&File{AppID: app.ID, PID: 0, Name: ""}).Assign(
		&File{IsDir: 1, UID: bson.NewObjectId().Hex()}).FirstOrCreate(file).Error
	return file, err
}

// CreateFileFromReader is used to create a file from reader.
func CreateFileFromReader(app *App, path string, reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) (*File, error) {
	var (
		object    *Object
		err       error
		file      *File
		parentDir *File
		dirPrefix = filepath.Dir(path)
		fileName  = filepath.Base(path)
	)

	if f, err := FindFileByPath(app, path, db); err == nil && f.ID > 0 {
		return nil, ErrFileExisted
	}

	if parentDir, err = CreateOrGetLastDirectory(app, dirPrefix, db); err != nil {
		return nil, err
	}

	if object, err = CreateObjectFromReader(reader, rootPath, db); err != nil {
		return nil, err
	}

	file = &File{
		UID:      bson.NewObjectId().Hex(),
		PID:      parentDir.ID,
		AppID:    app.ID,
		ObjectID: object.ID,
		Size:     object.Size,
		Name:     fileName,
		Ext:      strings.TrimPrefix(filepath.Ext(fileName), "."),
		Hidden:   hidden,
		Object:   *object,
	}

	if err = db.Save(file).Error; err != nil {
		return nil, err
	}

	if err = parentDir.UpdateParentSize(object.Size, db); err != nil {
		return file, err
	}

	return file, err
}

// FindFileByPath is used to find a file by the specify path
func FindFileByPath(app *App, path string, db *gorm.DB) (*File, error) {
	var (
		err    error
		parent = &File{}
		parts  = strings.Split(strings.Trim(strings.TrimSpace(path), "/"), string(os.PathSeparator))
	)

	if parent, err = CreateOrGetRootPath(app, db); err != nil {
		return nil, err
	}

	for _, part := range parts {
		var file = &File{}
		if err = db.Where("appId = ? and pid = ? and name = ?", app.ID, parent.ID, part).Find(file).Error; err != nil {
			return nil, err
		}
		parent = file
	}
	return parent, nil
}
