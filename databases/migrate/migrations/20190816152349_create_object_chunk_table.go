//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateObjectChunkTable20190816152349{})
}

// CreateObjectChunkTable20190816152349 represent some database operate
type CreateObjectChunkTable20190816152349 struct{}

// Name represent operate name, it's unique
func (c *CreateObjectChunkTable20190816152349) Name() string {
	return "create_object_chunk_table_20190816152349"
}

// Up is executed in upgrading
func (c *CreateObjectChunkTable20190816152349) Up(db *gorm.DB) error {
	// execute when upgrade database
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS object_chunk (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  objectId BIGINT(20) NOT NULL,
		  chunkId BIGINT(20) NOT NULL,
		  hashState text NULL,
		  number BIGINT(20) NOT NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
          updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  PRIMARY KEY (id),
		  UNIQUE INDEX object_chunk_no_uq (objectId, chunkId, number),
          KEY objectId_idx (objectId),
          KEY chunkId_idx (chunkId)
		)ENGINE = InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci
	`).Error
}

// Down is executed in downgrading
func (c *CreateObjectChunkTable20190816152349) Down(db *gorm.DB) error {
	// execute when rollback database
	return db.DropTableIfExists("object_chunk").Error
}
