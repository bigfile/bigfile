//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import "time"

// ChunkSize represent chunk size
const ChunkSize = 1 << 20

// Chunk represents every chunk of file
type Chunk struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Size      int       `gorm:"type:int;column:size"`
	Hash      string    `gorm:"type:CHAR(64) NOT NULL;UNIQUE;column:hash"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
}

// TableName represent table name
func (c Chunk) TableName() string {
	return "chunks"
}
