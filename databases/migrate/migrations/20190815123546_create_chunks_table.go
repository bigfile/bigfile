//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateChunksTable20190815123546{})
}

// CreateChunksTable20190815123546 represent some database operate
type CreateChunksTable20190815123546 struct{}

// Name represent operate name, it's unique
func (c *CreateChunksTable20190815123546) Name() string {
	return "create_chunks_table_20190815123546"
}

// Up is executed in upgrading
func (c *CreateChunksTable20190815123546) Up(db *gorm.DB) error {
	// execute when upgrade database
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS chunks (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  size INT UNSIGNED NOT NULL,
		  hash CHAR(64) NOT NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  PRIMARY KEY (id),
		  UNIQUE INDEX hash_UNIQUE (hash ASC))
		ENGINE = InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4
    `).Error
}

// Down is executed in downgrading
func (c *CreateChunksTable20190815123546) Down(db *gorm.DB) error {
	// execute when rollback database
	return db.DropTableIfExists("chunks").Error
}
