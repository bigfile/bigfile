//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateAppsTable20190801213442{})
}

// CreateAppsTable20190801213442 represent some database operate
type CreateAppsTable20190801213442 struct{}

// Name represent operate name, it's unique
func (c *CreateAppsTable20190801213442) Name() string {
	return "create_apps_table_20190801213442"
}

// Up is executed in upgrading
func (c *CreateAppsTable20190801213442) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS apps (
	  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	  secret CHAR(32) NOT NULL COMMENT 'Application Secret',
	  name VARCHAR(100) NOT NULL COMMENT 'Application Name',
	  note VARCHAR(500) NULL COMMENT 'Application Note',
	  createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  updatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  deletedAt TIMESTAMP NULL,
	  PRIMARY KEY (id),
	  INDEX deleted_at_idx (deletedAt)
	)`).Error
}

// Down is executed in downgrading
func (c *CreateAppsTable20190801213442) Down(db *gorm.DB) error {
	return db.DropTableIfExists("apps").Error
}
