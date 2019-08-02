//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package main provides a database migrate cmd, for keeping in
// step with team members, you can find help by this command:
//
// 		go run cmd/migrate/migrate.go  --help
//
// create migration files:
// 		go run cmd/migrate/migrate.go create alter_users_table
// execute upgrade command:
//		go run cmd/migrate/migrate.go upgrade
// execute rollback command:
//	     go run cmd/migrate/migrate.go rollback
package main

import (
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/migrate"
	_ "github.com/bigfile/bigfile/databases/migrate/migrations"
	"github.com/jinzhu/gorm"
	"gopkg.in/urfave/cli.v2"
)

var migrateTpl = `
//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package migrations

import (
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&{{ .StructureName }}{})
}

// {{ .StructureName }} represent some database operate
type {{ .StructureName }} struct{}

// Name represent operate name, it's unique
func (c *{{ .StructureName }}) Name() string {
	return "{{ .Name }}"
}

// Up is executed in upgrading 
func (c *{{ .StructureName }}) Up(db *gorm.DB) error {
	// execute when upgrade database
	return nil
}

// Down is executed in downgrading
func (c *{{ .StructureName }}) Down(db *gorm.DB) error {
	// execute when rollback database
	return nil
}

`

func main() {

	var (
		configurator = &config.Configurator{}
		app          cli.App
		connection   *gorm.DB
		err          error
	)

	app = cli.App{
		Name:     "migrate",
		Version:  "0.1.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "bigfile team",
				Email: "bigfile@gmail.com",
			},
		},
		Copyright: "copyright 2019 The bigfile authors",
		Usage:     "create, upgrade and rollback database",
		UsageText: "migrate [global options] command [command options] [arguments...]",
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "create migration files",
				UsageText: "migrate create [command options] migration_file_name",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "migration files path",
						Value:   "databases/migrate/migrations",
					},
				},
				Action: func(context *cli.Context) error {
					if context.NArg() < 1 {
						return fmt.Errorf("migration file name is empty")
					}
					var (
						path          = context.String("path")
						stat          os.FileInfo
						fileName      = context.Args().Get(0)
						structureName string
						timestamp     string
						tpl           *template.Template
						variables     struct {
							StructureName string
							Name          string
						}
						migrationFile *os.File
					)

					if stat, err = os.Stat(path); err != nil {
						return err
					} else if !stat.IsDir() {
						return fmt.Errorf("path must be a directory")
					}

					timestamp = time.Now().Format("20060102150405")
					structureName = strings.Replace(fileName, "_", " ", -1)
					structureName = strings.Title(structureName)
					structureName = strings.Replace(structureName, " ", "", -1)
					variables.StructureName = fmt.Sprintf("%s%s", structureName, timestamp)
					variables.Name = fmt.Sprintf("%s_%s", fileName, timestamp)

					if tpl, err = template.New("migrate tpl").Parse(migrateTpl); err != nil {
						return err
					}

					fileName = fmt.Sprintf("%s/%s_%s.go", strings.TrimRight(path, "/"), timestamp, fileName)
					if migrationFile, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
						return err
					}
					defer migrationFile.Close()

					if err = tpl.Execute(migrationFile, variables); err != nil {
						return err
					}

					fmt.Printf("\033[32mCreate migration file: %s\033[0m\n", fileName)

					return nil
				},
			},
			{
				Name:      "upgrade",
				Usage:     "execute migrate to upgrade database",
				UsageText: "migrate upgrade",
				Action: func(context *cli.Context) error {
					migrate.DefaultMC.Upgrade()
					fmt.Println("\033[32mMigrate: done!\033[0m")
					return nil
				},
				Before: func(context *cli.Context) error {
					connection, err = databases.NewConnection(&configurator.Database, true)
					migrate.DefaultMC.SetConnection(connection)
					return err
				},
			},
			{
				Name:      "rollback",
				Usage:     "rollback database",
				UsageText: "migrate rollback [command options]",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "step",
						Aliases: []string{"s"},
						Usage:   "rollback step",
						Value:   1,
					},
				},
				Action: func(context *cli.Context) error {
					migrate.DefaultMC.Rollback(uint(context.Int("step")))
					fmt.Println("\033[31mRollback: done!\033[0m")
					return nil
				},
				Before: func(context *cli.Context) error {
					connection, err = databases.NewConnection(&configurator.Database, true)
					migrate.DefaultMC.SetConnection(connection)
					return err
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config,c",
				Usage:   "system config file",
				Value:   "bigfile.yaml",
				Aliases: []string{"c"},
				EnvVars: []string{"ENV_BIGFILE_CONFIG"},
			},
		},
		Before: func(context *cli.Context) error {
			return config.ParseConfigFile(context.String("config"), configurator)
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
