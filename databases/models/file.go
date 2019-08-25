//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import (
	"errors"
	"fmt"
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
	// ErrFileExisted represent that the path has been occupied
	ErrFileExisted = errors.New("file has already existed")
	// ErrOverwriteDir represent that try to overwrite some directory
	ErrOverwriteDir = errors.New("directory can't be overwritten")
	// ErrAppendToDir represent that try to append content to directory
	ErrAppendToDir = errors.New("can't append data to directory")
	// ErrReadDir represent that can't read data from directory, only file
	ErrReadDir = errors.New("can't read a directory")
	// ErrAccessDenied represent a file can't be accessed by some tokens
	ErrAccessDenied = errors.New("file can't be accessed by some tokens")
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

	Object    Object    `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	App       App       `gorm:"foreignkey:appId;association_autoupdate:false;association_autocreate:false"`
	Parent    *File     `gorm:"foreignkey:id;association_foreignkey:pid;association_autoupdate:false;association_autocreate:false"`
	Histories []History `gorm:"foreignkey:fileId;association_autoupdate:false;association_autocreate:false"`
}

func pathCacheKey(app *App, path string) string {
	return fmt.Sprintf("%d-%s", app.ID, path)
}

// TableName represent the name of files table
func (f *File) TableName() string {
	return "files"
}

// CanBeAccessedByToken represent whether the file can be accessed by the token
func (f *File) CanBeAccessedByToken(token *Token, db *gorm.DB) error {
	var (
		err  error
		path string
	)
	if path, err = f.Path(db); err != nil {
		return err
	}
	if !strings.HasPrefix(path, token.Path) {
		return ErrAccessDenied
	}
	return nil
}

// Reader is used to get reader that continues to read data from underlying
// chunk until io.EOF
func (f *File) Reader(rootPath *string, db *gorm.DB) (io.Reader, error) {
	if f.IsDir == 1 {
		return nil, ErrReadDir
	}
	var (
		err error
	)
	if len(f.Object.Chunks) == 0 {
		if err = db.Preload("Chunks").Find(&f.Object).Error; err != nil {
			return nil, err
		}
	}
	return (&f.Object).Reader(rootPath)
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

func (f *File) createHistory(objectId uint64, path string, db *gorm.DB) error {
	return db.Save(&History{ObjectID: objectId, FileID: f.ID, Path: path}).Error
}

// OverWriteFromReader is used to overwrite the object
func (f *File) OverWriteFromReader(reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) error {

	if f.IsDir == 1 {
		return ErrOverwriteDir
	}

	var (
		err      error
		path     string
		object   *Object
		sizeDiff int
	)

	if path, err = f.Path(db); err != nil {
		return err
	}

	if err := f.createHistory(f.ObjectID, path, db); err != nil {
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

	if err = db.Model(f).Update(map[string]interface{}{
		"objectId": object.ID,
		"hidden":   hidden,
		"size":     f.Size,
	}).Error; err != nil {
		return err
	}

	if err = db.Preload("Parent").Preload("App").Find(f).Error; err != nil {
		return err
	}

	return f.Parent.UpdateParentSize(sizeDiff, db)
}

func (f *File) mustPath(db *gorm.DB) string {
	var (
		err  error
		path string
	)
	if path, err = f.Path(db); err != nil {
		panic(err)
	}
	return path
}

// MoveTo move file to another path, the input path must be complete and new path.
// if the input path ios the same as the previous path, nothing changes.
func (f *File) MoveTo(newPath string, db *gorm.DB) error {
	var (
		err             error
		newPathDir      = filepath.Dir(newPath)
		newPathDirFile  *File
		newPathFileName = filepath.Base(newPath)
		newPathExt      = strings.TrimPrefix(filepath.Ext(newPathFileName), ".")
		previousPath    string
	)

	if previousPath, err = f.Path(db); err != nil {
		return err
	}

	if previousPath == newPath {
		return nil
	}

	if f.App.ID == 0 {
		if err = db.Preload("App").Find(f).Error; err != nil {
			return err
		}
	}

	if newPathDirFile, err = CreateOrGetLastDirectory(&f.App, newPathDir, db); err != nil {
		return err
	}

	if f.IsDir == 0 {
		if err = f.createHistory(f.ObjectID, previousPath, db); err != nil {
			return err
		}
	}

	f.Name = newPathFileName
	f.Ext = newPathExt

	defer func() {
		pathToFileCache.Delete(previousPath)
		_ = pathToFileCache.Add(pathCacheKey(&f.App, f.mustPath(db)), f, time.Hour*48)
	}()

	// only change the file name, still is in the same directory
	if newPathDirFile.ID == f.PID {
		return db.Save(f).Error
	}

	if f.Parent.ID == 0 {
		if err = db.Preload("Parent").Find(f).Error; err != nil {
			return err
		}
	}

	if err = newPathDirFile.UpdateParentSize(f.Size, db); err != nil {
		return err
	}

	if err = f.Parent.UpdateParentSize(-f.Size, db); err != nil {
		return err
	}
	f.PID = newPathDirFile.ID
	f.Parent = newPathDirFile
	return db.Save(f).Error
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
	if err = db.Preload("Object").Preload("Parent").Preload("App").Find(f).Error; err != nil {
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
func CreateOrGetLastDirectory(app *App, parentDirs string, db *gorm.DB) (*File, error) {
	var (
		parent *File
		err    error
		parts  = strings.Split(strings.Trim(strings.TrimSpace(parentDirs), "/"), string(os.PathSeparator))
	)

	if parent, err = CreateOrGetRootPath(app, db); err != nil {
		return nil, err
	}

	for _, part := range parts {
		file := &File{}
		if err = db.Where("appId = ? and pid = ? and name = ?", app.ID, parent.ID, part).Find(file).Error; err != nil {
			if !util.IsRecordNotFound(err) {
				return nil, err
			}
			file.AppID = app.ID
			file.PID = parent.ID
			file.Name = part
			file.IsDir = 1
			file.UID = bson.NewObjectId().Hex()
			if err = db.Save(file).Error; err != nil {
				return nil, err
			}
		}
		parent = file
	}
	parent.App = *app
	return parent, nil
}

// CreateOrGetRootPath is used to create or get root directory
func CreateOrGetRootPath(app *App, db *gorm.DB) (*File, error) {
	var (
		file = &File{}
		err  error
	)
	err = db.Where("appId = ? & pid = 0 and name = ''", app.ID).Find(file).Error
	file.App = *app
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
		App:      *app,
		Parent:   parentDir,
	}

	if err = db.Save(file).Error; err != nil {
		return nil, err
	}

	if err = parentDir.UpdateParentSize(object.Size, db); err != nil {
		return file, err
	}

	return file, err
}

// FindFileByUID is used to find a file by uid
func FindFileByUID(uid string, trashed bool, db *gorm.DB) (*File, error) {
	var (
		file = &File{}
		err  error
	)
	if trashed {
		db = db.Unscoped()
	}
	if err = db.Where("uid = ?", uid).Find(file).Error; err != nil {
		return file, err
	}
	return file, nil
}

// FindFileByPath is used to find a file by the specify path
func FindFileByPath(app *App, path string, db *gorm.DB) (*File, error) {
	var cacheKey = pathCacheKey(app, path)

	if fileValue, ok := pathToFileCache.Get(cacheKey); ok {
		file := fileValue.(*File)
		if err := db.Where("id = ?", file.ID).Find(file).Error; err != nil {
			file.App = *app
			return file, nil
		}
	}

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
	parent.App = *app

	_ = pathToFileCache.Add(cacheKey, parent, time.Hour*48)

	return parent, nil
}
