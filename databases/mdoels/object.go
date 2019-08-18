//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package models

import (
	"time"

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

	Chunks       []Chunk       `gorm:"many2many:object_chunk;association_jointable_foreignkey:chunkId;jointable_foreignkey:objectId"`
	ObjectChunks []ObjectChunk `gorm:"foreignkey:objectId"`
}

// TableName represent the db table name
func (o Object) TableName() string {
	return "objects"
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
	err := db.Preload("ObjectChunks", func(db *gorm.DB) *gorm.DB {
		return db.Order("object_chunk.id desc").Limit(1)
	}).Find(o).Error
	if len(o.ObjectChunks) > 0 {
		number = o.ObjectChunks[0].Number
	}
	return number, err
}
