//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package migrations hold the migration file for operate database
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
	CREATE TABLE apps (
	  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
	  uid char(32) not null comment 'Application unique id',
	  secret char(32) NOT NULL COMMENT 'Application Secret',
	  name varchar(100) NOT NULL COMMENT 'Application Name',
	  note varchar(500) DEFAULT NULL COMMENT 'Application Note',
	  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
	  deletedAt timestamp(6) NULL DEFAULT NULL,
	  PRIMARY KEY (id),
	  UNIQUE INDEX uid_uq_idx (uid),
	  KEY deleted_at_idx (deletedAt)
	) ENGINE=InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

// Down is executed in downgrading
func (c *CreateAppsTable20190801213442) Down(db *gorm.DB) error {
	return db.DropTableIfExists("apps").Error
}
