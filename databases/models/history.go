//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package models

import "time"

// History represent the overwrite history of object. By this, we
// can easily find kinds of versions of the object.
type History struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	ObjectID  uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:objectId"`
	FileID    uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:fileId"`
	Path      string    `gorm:"type:tinyint;column:path"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
}

// TableName represent the name of history table
func (h *History) TableName() string {
	return "histories"
}
