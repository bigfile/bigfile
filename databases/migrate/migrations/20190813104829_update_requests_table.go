//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&UpdateRequestsTable20190813104829{})
}

// UpdateRequestsTable20190813104829 represent some database operate
type UpdateRequestsTable20190813104829 struct{}

// Name represent operate name, it's unique
func (c *UpdateRequestsTable20190813104829) Name() string {
	return "update_requests_table_20190813104829"
}

// Up is executed in upgrading
func (c *UpdateRequestsTable20190813104829) Up(db *gorm.DB) error {
	return db.Exec(`
	alter table requests 
		add column nonce char(48) default null after appId,
		add column requestHeader text after requestBody,
		add index appId_idx (appId)
	`).Error
}

// Down is executed in downgrading
func (c *UpdateRequestsTable20190813104829) Down(db *gorm.DB) error {
	// execute when rollback database
	return db.Exec(`
	alter table requests 
		drop index appId_idx,
		drop column requestHeader, 
		drop column nonce
	`).Error
}
