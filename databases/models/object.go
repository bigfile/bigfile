//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
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

	Files        []File        `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	Chunks       []Chunk       `gorm:"many2many:object_chunk;association_jointable_foreignkey:chunkId;jointable_foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	ObjectChunks []ObjectChunk `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	Histories    []History     `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
}

// TableName represent the db table name
func (o Object) TableName() string {
	return "objects"
}

// FileCountWithTrashed count the files they are associated with this object
func (o *Object) FileCountWithTrashed(db *gorm.DB) int {
	return db.Unscoped().Model(o).Association("Files").Count()
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
	var (
		oc     *ObjectChunk
		err    error
		number int
	)
	oc, err = o.LastObjectChunk(db)
	if oc != nil {
		return oc.Number, nil
	}
	return number, err
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
func (o *Object) AppendFromReader(reader io.Reader, rootPath *string, db *gorm.DB) (*Object, int, error) {
	var (
		err              error
		lastOc           *ObjectChunk
		object           *Object
		lastChunk        *Chunk
		stateHash        hash.Hash
		readerContent    []byte
		completeHashStr  string
		readerContentLen int
	)
	if readerContent, err = ioutil.ReadAll(reader); err != nil {
		return o, 0, err
	}

	if readerContentLen = len(readerContent); readerContentLen <= 0 {
		return o, 0, nil
	}

	if lastOc, err = o.LastObjectChunk(db); err != nil {
		return o, 0, err
	}

	if stateHash, err = sha2562.NewHashWithStateText(*lastOc.HashState); err != nil {
		return o, 0, err
	}
	if _, err = stateHash.Write(readerContent); err != nil {
		return o, 0, err
	}

	completeHashStr = hex.EncodeToString(stateHash.Sum(nil))

	if object, err = FindObjectByHash(completeHashStr, db); err == nil && object != nil {
		return object, len(readerContent), nil
	}

	object = &Object{
		Size: o.Size + len(readerContent),
		Hash: completeHashStr,
	}
	stateHash, _ = sha2562.NewHashWithStateText(*lastOc.HashState)
	if err = db.Where("objectId = ?", o.ID).Find(&object.ObjectChunks).Error; err != nil {
		return o, 0, err
	}
	// determine if we need to copy the object
	if o.FileCountWithTrashed(db)+db.Model(o).Association("Histories").Count() <= 1 {
		object.ID = o.ID
		object.CreatedAt = o.CreatedAt
		object.UpdatedAt = o.UpdatedAt
	} else {
		// copy the object chunk
		for index := range object.ObjectChunks {
			object.ObjectChunks[index].ID = 0
		}
	}

	// get the last chunk of object, determine if we need to complete the last chunk
	if lastChunk, err = o.LastChunk(db); err != nil {
		return o, 0, err
	}

	if lackSize := ChunkSize - lastChunk.Size; lackSize > 0 {
		var (
			err       error
			chunk     *Chunk
			hashState string
		)

		if lackSize > len(readerContent) {
			lackSize = len(readerContent)
		}

		if chunk, _, err = lastChunk.AppendBytes(readerContent[:lackSize], rootPath, db); err != nil {
			return o, 0, err
		}
		if chunk.ID != lastChunk.ID {
			object.ObjectChunks[len(object.ObjectChunks)-1].ChunkID = chunk.ID
		}
		if _, err := stateHash.Write(readerContent[:lackSize]); err != nil {
			return o, 0, err
		}
		if hashState, err = sha2562.GetHashStateText(stateHash); err != nil {
			return o, 0, err
		}
		object.ObjectChunks[len(object.ObjectChunks)-1].HashState = &hashState
		readerContent = readerContent[lackSize:]
	}
	if err = appendContentToObject(
		object, object.ObjectChunks, readerContent, lastOc.Number, stateHash, rootPath, db); err != nil {
		return o, 0, nil
	}

	return object, readerContentLen, nil
}

// Reader is used to implement io.Reader
func (o *Object) Reader(rootPath *string) (io.Reader, error) {
	return NewObjectReader(o, rootPath)
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
		err           error
		object        *Object
		sha256Hash    = sha256.New()
		contentHash   string
		readerContent []byte
	)

	if readerContent, err = ioutil.ReadAll(reader); err != nil {
		return nil, err
	}

	if len(readerContent) == 0 {
		return CreateEmptyObject(rootPath, db)
	}

	if contentHash, err = util.Sha256Hash2String(readerContent); err != nil {
		return nil, err
	}

	if object, err = FindObjectByHash(contentHash, db); err == nil && object != nil {
		return object, nil
	}

	object = &Object{
		Size: len(readerContent),
		Hash: contentHash,
	}

	return object, appendContentToObject(object, nil, readerContent, 0, sha256Hash, rootPath, db)
}

// CreateEmptyObject is used to create an empty object
func CreateEmptyObject(rootPath *string, db *gorm.DB) (*Object, error) {
	var (
		h                = sha256.New()
		err              error
		chunk            *Chunk
		object           *Object
		hashState        string
		emptyContentHash = hex.EncodeToString(h.Sum(nil))
	)

	if object, err = FindObjectByHash(emptyContentHash, db); err == nil && object != nil {
		return object, nil
	}

	if chunk, err = CreateEmptyContentChunk(rootPath, db); err != nil {
		return nil, err
	}

	if hashState, err = sha2562.GetHashStateText(h); err != nil {
		return nil, err
	}

	object = &Object{
		Size: 0,
		Hash: emptyContentHash,
		ObjectChunks: []ObjectChunk{
			{
				ChunkID:   chunk.ID,
				Number:    1,
				HashState: &hashState,
			},
		},
	}

	return object, db.Set("gorm:association_autocreate", true).Save(object).Error
}

func appendContentToObject(obj *Object, oc []ObjectChunk, readerContent []byte, index int, hash hash.Hash, rootPath *string, db *gorm.DB) error {
	var (
		err        error
		contentBuf = bytes.NewReader(readerContent)
	)

	for i := index; contentBuf.Len() > 0; i++ {
		var (
			chunk     *Chunk
			content   = make([]byte, ChunkSize)
			readLen   int
			hashState string
		)
		if readLen, err = contentBuf.Read(content); err != nil {
			return err
		}
		if chunk, err = CreateChunkFromBytes(content[:readLen], rootPath, db); err != nil {
			return err
		}
		if _, err := hash.Write(content[:readLen]); err != nil {
			return err
		}
		if hashState, err = sha2562.GetHashStateText(hash); err != nil {
			return err
		}
		oc = append(oc, ObjectChunk{
			ChunkID:   chunk.ID,
			Number:    i + 1,
			HashState: &hashState,
		})
	}

	if err = db.Save(obj).Error; err != nil {
		return err
	}

	for _, objectChunk := range oc {
		objectChunk.ObjectID = obj.ID
		if err := db.Save(&objectChunk).Error; err != nil {
			return err
		}
	}

	return nil
}
