//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"math"
	"time"

	sha2562 "github.com/bigfile/bigfile/internal/sha256"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/jinzhu/gorm"
)

// Object represent a documentation that is correspond to system
// An object has many chunks, it's saved in disk by chunk. But,
// a file is a documentation that is correspond to user.
type Object struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Size      int       `gorm:"type:int;column:size"`
	Hash      string    `gorm:"type:CHAR(64) NOT NULL;UNIQUE;column:hash"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`

	Files        []File        `gorm:"foreignkey:objectId"`
	Chunks       []Chunk       `gorm:"many2many:object_chunk;association_jointable_foreignkey:chunkId;jointable_foreignkey:objectId"`
	ObjectChunks []ObjectChunk `gorm:"foreignkey:objectId"`
}

// TableName represent the db table name
func (o Object) TableName() string {
	return "objects"
}

// FileCount count the files they are associated with this object
func (o *Object) FileCount(db *gorm.DB) int {
	return db.Model(o).Association("Files").Count()
}

// ChunkCount count the chunks of object and return
func (o *Object) ChunkCount(db *gorm.DB) int {
	return db.Model(o).Association("Chunks").Count()
}

// LastChunk return the last chunk of object
func (o *Object) LastChunk(db *gorm.DB) (*Chunk, error) {
	var (
		joinObjectChunk = "join object_chunk on object_chunk.chunkId = chunks.id and object_chunk.objectId = ?"
		chunk           = &Chunk{}
		err             error
	)
	err = db.Joins(joinObjectChunk, o.ID).Order("chunks.id desc").First(chunk).Error
	return chunk, err
}

// LastChunkNumber is used to return the last chunk number
// chunk number starts from 1, 0 represent no chunks
func (o *Object) LastChunkNumber(db *gorm.DB) (int, error) {
	var number int
	if oc, err := o.LastObjectChunk(db); err != nil {
		return 0, err
	} else {
		if oc != nil {
			return oc.Number, nil
		}
	}
	return number, nil
}

// LastObjectChunk return the middle value between chunk and object
func (o *Object) LastObjectChunk(db *gorm.DB) (*ObjectChunk, error) {
	err := db.Preload("ObjectChunks", func(db *gorm.DB) *gorm.DB {
		return db.Order("object_chunk.id desc").Limit(1)
	}).Find(o).Error
	if len(o.ObjectChunks) == 0 {
		return nil, err
	}
	return &o.ObjectChunks[0], nil
}

// AppendFromReader will append content from reader to object
func (o *Object) AppendFromReader(reader io.Reader, db *gorm.DB) (*Object, int, error) {
	var (
		allContent      []byte
		allContentLen   int
		err             error
		lastOc          *ObjectChunk
		lastChunk       *Chunk
		stateHash       hash.Hash
		completeHashStr string
		object          *Object
	)
	if allContent, err = ioutil.ReadAll(reader); err != nil {
		return o, 0, err
	}
	allContentLen = len(allContent)

	if lastOc, err = o.LastObjectChunk(db); err != nil {
		return o, 0, err
	} else if lastOc == nil {
		return o, 0, errors.New("unexpected error happened, object must have some chunks")
	}

	if stateHash, err = sha2562.NewHashWithStateText(*lastOc.HashState); err != nil {
		return o, 0, err
	}
	if _, err = stateHash.Write(allContent); err != nil {
		return o, 0, err
	}
	completeHashStr = hex.EncodeToString(stateHash.Sum(nil))
	if object, err = FindObjectByHash(completeHashStr, db); err == nil && object != nil {
		return object, len(allContent), nil
	}
	// determine if we need to copy the object

	// get the last chunk of object, determine if we need to complete
	// the last chunk
	if lastChunk, err = o.LastChunk(db); err != nil {
		return o, 0, err
	}

	_ = allContentLen
	_ = lastChunk

	// append the rest of content

	return &Object{}, 0, nil
}

// FindObjectByHash will find object by the specify hash
func FindObjectByHash(h string, db *gorm.DB) (*Object, error) {
	var object Object
	var err = db.Where("hash = ?", h).First(&object).Error
	return &object, err
}

// CreateObjectFromReader reads data from reader to create an object
func CreateObjectFromReader(reader io.Reader, rootPath *string, db *gorm.DB) (*Object, error) {
	var (
		oc         []*ObjectChunk
		sha256Hash = sha256.New()
		allContent []byte
		err        error
	)

	if allContent, err = ioutil.ReadAll(reader); err != nil {
		return nil, err
	}

	if contentHash, err := util.Sha256Hash2String(allContent); err != nil {
		return nil, err
	} else {
		if object, err := FindObjectByHash(contentHash, db); err == nil && object != nil {
			return object, nil
		}
	}

	maxChunkNumber := int(math.Ceil(float64(len(allContent)) / float64(ChunkSize)))
	for i := 0; i < maxChunkNumber; i++ {

		end := (i + 1) * ChunkSize
		if end > len(allContent) {
			end = len(allContent)
		}

		var content = allContent[i*ChunkSize : end]
		if chunk, err := CreateChunkFromBytes(content, rootPath, db); err != nil {
			return nil, err
		} else {
			if _, err := sha256Hash.Write(content); err != nil {
				return nil, err
			}
			hashState, err := sha2562.GetHashStateText(sha256Hash)
			if err != nil {
				return nil, err
			}
			oc = append(oc, &ObjectChunk{
				ChunkID:   chunk.ID,
				Number:    i + 1,
				HashState: &hashState,
			})
		}
	}

	// for atomicity, the following code should be run in a transaction
	if !isTesting {
		db = db.Begin()
		defer func() {
			if rErr := recover(); rErr != nil || err != nil {
				db.Rollback()
			}
		}()
	}
	object := &Object{
		Size: len(allContent),
		Hash: hex.EncodeToString(sha256Hash.Sum(nil)),
	}
	if err = db.Save(object).Error; err != nil {
		return nil, err
	}

	for _, objectChunk := range oc {
		objectChunk.ObjectID = object.ID
		if err := db.Save(objectChunk).Error; err != nil {
			return nil, err
		}
	}

	if !isTesting {
		return object, db.Commit().Error
	}

	return object, nil
}
