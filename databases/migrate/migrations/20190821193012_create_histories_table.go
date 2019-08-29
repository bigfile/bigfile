//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateHistoriesTable20190821193012{})
}

// CreateHistoriesTable20190821193012 represent some database operate
type CreateHistoriesTable20190821193012 struct{}

// Name represent operate name, it's unique
func (c *CreateHistoriesTable20190821193012) Name() string {
	return "create_histories_table_20190821193012"
}

// Up is executed in upgrading
func (c *CreateHistoriesTable20190821193012) Up(db *gorm.DB) error {
	// execute when upgrade database
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS histories (
	  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	  fileId BIGINT(20) UNSIGNED NOT NULL,
	  objectId BIGINT(20) UNSIGNED NOT NULL,
	  path VARCHAR(1000) NOT NULL,
	  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	  PRIMARY KEY (id))
	ENGINE = InnoDB`).Error
}

// Down is executed in downgrading
func (c *CreateHistoriesTable20190821193012) Down(db *gorm.DB) error {
	// execute when rollback database
	return db.DropTableIfExists("histories").Error
}
