//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateFilesTable20190820111006{})
}

// CreateFilesTable20190820111006 represent some database operate
type CreateFilesTable20190820111006 struct{}

// Name represent operate name, it's unique
func (c *CreateFilesTable20190820111006) Name() string {
	return "create_files_table_20190820111006"
}

// Up is executed in upgrading
func (c *CreateFilesTable20190820111006) Up(db *gorm.DB) error {
	// execute when upgrade database
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  appId BIGINT(20) UNSIGNED NOT NULL default 0,
		  pid BIGINT(20) UNSIGNED NOT NULL DEFAULT 0,
		  uid CHAR(32) NOT NULL,
		  name VARCHAR(255) NOT NULL default '',
		  ext VARCHAR(255) NOT NULL default '',
		  objectId BIGINT(20) UNSIGNED NOT NULL default 0,
		  size BIGINT(20) UNSIGNED NOT NULL default 0,
		  isDir TINYINT UNSIGNED NOT NULL DEFAULT 0,
		  downloadCount BIGINT(20) UNSIGNED NOT NULL DEFAULT 0,
		  hidden TINYINT UNSIGNED NOT NULL DEFAULT 0,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  deletedAt timestamp(6) NULL DEFAULT NULL,
		  PRIMARY KEY (id),
		  KEY objectId_idx (objectId),
		  KEY appId_idx (appId),
		  KEY deleted_at_idx (deletedAt),
          UNIQUE appId_pid_name_unique (appId, pid, name),
		  UNIQUE INDEX uid_UNIQUE (uid ASC))
		ENGINE = InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci
	`).Error
}

// Down is executed in downgrading
func (c *CreateFilesTable20190820111006) Down(db *gorm.DB) error {
	// execute when rollback database
	return db.DropTableIfExists("files").Error
}
