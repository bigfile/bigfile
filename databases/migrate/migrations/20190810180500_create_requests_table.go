//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateRequestsTable20190810180500{})
}

// CreateRequestsTable20190810180500 represent some database operate
type CreateRequestsTable20190810180500 struct{}

// Name represent operate name, it's unique
func (c *CreateRequestsTable20190810180500) Name() string {
	return "create_requests_table_20190810180500"
}

// Up is executed in upgrading
func (c *CreateRequestsTable20190810180500) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS requests (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  protocol CHAR(10) NOT NULL,
		  appId BIGINT(20) NULL DEFAULT NULL,
		  token CHAR(32) NULL DEFAULT NULL,
		  ip CHAR(15) NULL DEFAULT NULL,
		  method CHAR(10) NULL DEFAULT NULL,
		  service VARCHAR(512) NULL DEFAULT NULL,
		  requestBody text NULL,
		  responseCode int NULL DEFAULT 200,
		  responseBody text NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  PRIMARY KEY (id)
      )ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

// Down is executed in downgrading
func (c *CreateRequestsTable20190810180500) Down(db *gorm.DB) error {
	return db.DropTableIfExists("requests").Error
}
