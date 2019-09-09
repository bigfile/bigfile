//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateObjectsTable20190816152340{})
}

// CreateObjectsTable20190816152340 represent some database operate
type CreateObjectsTable20190816152340 struct{}

// Name represent operate name, it's unique
func (c *CreateObjectsTable20190816152340) Name() string {
	return "create_objects_table_20190816152340"
}

// Up is executed in upgrading
func (c *CreateObjectsTable20190816152340) Up(db *gorm.DB) error {
	// execute when upgrade database
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS objects (
	  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	  size INT UNSIGNED NOT NULL,
	  hash CHAR(64) NOT NULL,
	  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
	  PRIMARY KEY (id),
	  UNIQUE INDEX hash_UNIQUE (hash ASC))
	ENGINE = InnoDB AUTO_INCREMENT=10000 DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci
	`).Error
}

// Down is executed in downgrading
func (c *CreateObjectsTable20190816152340) Down(db *gorm.DB) error {
	// execute when rollback database
	return db.DropTableIfExists("objects").Error
}
