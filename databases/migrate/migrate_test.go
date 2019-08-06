//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrate

import (
	"testing"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

type UserModelTest struct {
	Name string `gorm:"type:varchar(255);not null"`
}

func (u UserModelTest) TableName() string {
	return "user_model_test"
}

type CreateUserTableMigration20190725 struct{}

func (c CreateUserTableMigration20190725) Name() string {
	return "CreateUserTableMigration20190725"
}

func (c CreateUserTableMigration20190725) Up(conn *gorm.DB) error {
	return conn.CreateTable(&UserModelTest{}).Error
}

func (c CreateUserTableMigration20190725) Down(conn *gorm.DB) error {
	return conn.DropTable(&UserModelTest{}).Error
}

func getDBConnection(t *testing.T) *gorm.DB {

	dbConfig := &config.Database{
		Driver: "sqlite3",
		DBFile: ":memory:",
	}

	connection, err := databases.NewConnection(dbConfig)
	if err != nil {
		t.Fatal(err)
	}
	return connection
}

func TestMigrationCollection_CreateMigrateTable(t *testing.T) {
	connection := getDBConnection(t)
	DefaultMC.SetConnection(connection)
	DefaultMC.CreateMigrateTable()
	if !connection.HasTable(&MigrationModel{}) {
		t.Fatal("migrations table should been created already!")
	}
	defer connection.DropTableIfExists(&MigrationModel{})
}

func TestMigrationCollection_MaxBatch(t *testing.T) {
	connection := getDBConnection(t)
	DefaultMC.SetConnection(connection)
	if DefaultMC.MaxBatch() != 0 {
		t.Fatal("max batch of empty migrations table should be 0")
	}
	defer connection.DropTableIfExists(&MigrationModel{})
}

func TestMigrationCollection_Register(t *testing.T) {
	DefaultMC.Register(&CreateUserTableMigration20190725{})
	assert.Equal(t, 1, len(DefaultMC.migrations))
}

func TestMigrationCollection_Upgrade(t *testing.T) {
	connection := getDBConnection(t)
	DefaultMC.SetConnection(connection)
	DefaultMC.Register(&CreateUserTableMigration20190725{})
	DefaultMC.Upgrade()
	if !connection.HasTable(&UserModelTest{}) {
		t.Fatalf("table %s should be already existed in database\n", UserModelTest{}.TableName())
	}
	if DefaultMC.MaxBatch() != 1 {
		t.Fatal("max batch of migrations table should be 1")
	}
	defer connection.DropTableIfExists(&MigrationModel{})
	defer connection.DropTableIfExists(&UserModelTest{})
}

func TestMigrationCollection_Rollback(t *testing.T) {
	connection := getDBConnection(t)
	DefaultMC.SetConnection(connection)
	DefaultMC.Register(&CreateUserTableMigration20190725{})
	DefaultMC.Upgrade()
	DefaultMC.Rollback(1)
	if connection.HasTable(&UserModelTest{}) {
		t.Fatalf("table %s should be already deleted in database\n", UserModelTest{}.TableName())
	}
	if DefaultMC.MaxBatch() != 0 {
		t.Fatal("max batch of migrations table should be 0")
	}
	defer connection.DropTableIfExists(&MigrationModel{})
	defer connection.DropTableIfExists(&UserModelTest{})
}
