
//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateObjectChunkTable20190816152349{})
}

// CreateObjectChunkTable20190816152349 represent some database operate
type CreateObjectChunkTable20190816152349 struct{}

// Name represent operate name, it's unique
func (c *CreateObjectChunkTable20190816152349) Name() string {
	return "create_object_chunk_table_20190816152349"
}

// Up is executed in upgrading 
func (c *CreateObjectChunkTable20190816152349) Up(db *gorm.DB) error {
	// execute when upgrade database
	return nil
}

// Down is executed in downgrading
func (c *CreateObjectChunkTable20190816152349) Down(db *gorm.DB) error {
	// execute when rollback database
	return nil
}
