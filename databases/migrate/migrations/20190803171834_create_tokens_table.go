//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateTokensTable20190803171834{})
}

// CreateTokensTable20190803171834 represent some database operate
type CreateTokensTable20190803171834 struct{}

// Name represent operate name, it's unique
func (c *CreateTokensTable20190803171834) Name() string {
	return "create_tokens_table_20190803171834"
}

// Up is executed in upgrading
func (c *CreateTokensTable20190803171834) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS tokens (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  uid CHAR(32) NOT NULL,
		  appId BIGINT(20) UNSIGNED NOT NULL,
		  ip VARCHAR(1500) NULL DEFAULT NULL,
		  availableTimes INT NOT NULL DEFAULT -1,
		  readOnly TINYINT NOT NULL DEFAULT 0,
		  secret CHAR(32) NULL DEFAULT NULL,
		  path VARCHAR(1000) NOT NULL,
		  expiredAt timestamp(6) NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  deletedAt timestamp(6) NULL DEFAULT NULL,
		  PRIMARY KEY (id),
		  UNIQUE INDEX uid_uq_index (uid ASC),
		  KEY deleted_at_idx (deletedAt)
      )ENGINE=InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

// Down is executed in downgrading
func (c *CreateTokensTable20190803171834) Down(db *gorm.DB) error {
	return db.DropTableIfExists("tokens").Error
}
