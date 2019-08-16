
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
	return nil
}

// Down is executed in downgrading
func (c *CreateObjectsTable20190816152340) Down(db *gorm.DB) error {
	// execute when rollback database
	return nil
}
