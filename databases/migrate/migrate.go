//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package migrate mainly provides capacity to migrateBak and rollback
// database change automatically. keeping database structure is in
// consistency in team members.
package migrate

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Migrator is the interface that wraps Name, Up, Down Method.
//
// Name method should return unique migration name in one database.
// Up method will be called in database upgrade.
// Down method will be called in database rollback.
type Migrator interface {
	Name() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

// MigrationModel is used to generate migrations table in database
type MigrationModel struct {
	ID        uint   `gorm:"primary_key;AUTO_INCREMENT"`
	Migration string `gorm:"type:varchar(255);not null;UNIQUE_INDEX"`
	Batch     uint   `gorm:"type:int unsigned;not null"`
}

// TableName represent the name of migration table
func (m MigrationModel) TableName() string {
	return "migrations"
}

// MigrationCollection is a collection of migrations
type MigrationCollection struct {
	migrations map[string]Migrator
	connection *gorm.DB
}

// SetConnection is used to set the db connection of database that
// will be operated on.
func (m *MigrationCollection) SetConnection(db *gorm.DB) {
	m.connection = db
}

// Register will add migration to collection
func (m *MigrationCollection) Register(migrate Migrator) {
	if m.migrations == nil {
		m.migrations = make(map[string]Migrator)
	}
	m.migrations[migrate.Name()] = migrate
}

// CreateMigrateTable is used to create "migrations" table that
// includes all migration information.
func (m *MigrationCollection) CreateMigrateTable() {
	if !m.connection.HasTable(&MigrationModel{}) {
		m.connection.CreateTable(&MigrationModel{})
	}
}

// MaxBatch return the max batch number of migrations
func (m *MigrationCollection) MaxBatch() uint {
	m.CreateMigrateTable()
	var (
		batch struct {
			Batch uint
		}
		sql = fmt.Sprintf("select max(batch) as batch from %s", MigrationModel{}.TableName())
	)
	m.connection.Raw(sql).Scan(&batch)
	return batch.Batch
}

// Upgrade will apply new change to database, complete database
// structure upgrade.
func (m *MigrationCollection) Upgrade() {
	var (
		migrations         []MigrationModel
		currentBatchNumber uint
		finishedMigrations = make(map[string]struct{}, len(m.migrations))
	)

	currentBatchNumber = m.MaxBatch() + 1

	m.connection.Find(&migrations)
	for _, migration := range migrations {
		finishedMigrations[migration.Migration] = struct{}{}
	}

	for name, migration := range m.migrations {
		if _, ok := finishedMigrations[name]; !ok {
			if err := migration.Up(m.connection); err == nil {
				m.connection.Create(&MigrationModel{
					Migration: name,
					Batch:     currentBatchNumber,
				})
				fmt.Printf("\033[32mMigrate: %s\033[0m\n", name)
			} else {
				fmt.Printf("\033[31mMigrate: %s, %s\033[0m\n", name, err.Error())
				return
			}
		}
	}
}

// Rollback will move back last migrating version, step
// represents the number of fallback versions.
func (m *MigrationCollection) Rollback(step uint) {
	fallbackTo := m.MaxBatch() - step + 1
	var migrations []MigrationModel
	m.connection.Where("batch >= ?", fallbackTo).Order("id desc").Find(&migrations)
	for _, migration := range migrations {
		if err := m.migrations[migration.Migration].Down(m.connection); err == nil {
			m.connection.Delete(&migration)
			fmt.Printf("\033[31mRollback: %s\033[0m\n", migration.Migration)
		} else {
			fmt.Printf("\033[31mRollback: %s, %s\033[0m\n", migration.Migration, err.Error())
			return
		}
	}
}

// Refresh is used to refresh the whole database
func (m *MigrationCollection) Refresh() {
	m.Rollback(m.MaxBatch())
	m.Upgrade()
}

// DefaultMC  In fact, is not available, unless you set
// the database connection by SetConnection method
var DefaultMC = &MigrationCollection{}
