//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package app

import (
	"errors"
	"os"
	"strconv"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	models "github.com/bigfile/bigfile/databases/mdoels"
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/bigfile/bigfile/log"
	"github.com/jinzhu/gorm"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/urfave/cli.v2"
)

var (
	category   = "app"
	connection *gorm.DB
	err        error
	logger     = log.MustNewLogger(nil)
	before     = func(context *cli.Context) error {
		connection, err = databases.NewConnection(&config.DefaultConfig.Database)
		migrate.DefaultMC.SetConnection(connection)
		return err
	}
)

// Commands is used to new and delete app
var Commands = []*cli.Command{
	{
		Name:      "app:new",
		Category:  category,
		Usage:     "create new application",
		UsageText: "app:new [command options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "application name",
			},
			&cli.StringFlag{
				Name:  "note",
				Usage: "application description",
			},
		},
		Action: func(ctx *cli.Context) error {
			var (
				name = ctx.String("name")
				note = ctx.String("note")
			)
			if len(name) == 0 {
				logger.Error("name is empty \n")
				return nil
			}
			app, err := models.NewApp(name, &note, connection)
			if err != nil {
				return err
			}
			tableData := [][]string{
				{strconv.FormatUint(app.ID, 10), app.UID, app.Secret, app.Name, *app.Note, app.CreatedAt.Format("2006-01-02 15:04:05")},
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "UID", "Secret", "Name", "Note", "CreatedAt"})
			for _, v := range tableData {
				table.Append(v)
			}
			table.Render()
			return nil
		},
		Before: before,
	},
	{
		Name:      "app:delete",
		Category:  category,
		Usage:     "delete an application",
		UsageText: "app:delete [command options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "uid",
				Aliases: []string{"u"},
				Usage:   "application uid",
			},
			&cli.BoolFlag{
				Name:    "permanently",
				Aliases: []string{"p"},
				Usage:   "delete application permanently",
			},
		},
		Action: func(ctx *cli.Context) error {
			var (
				permanently = ctx.Bool("permanently")
				uid         = ctx.String("uid")
				err         error
			)
			if permanently {
				err = models.DeleteAppByUIDPermanently(uid, connection)
			} else {
				err = models.DeleteAppByUIDSoft(uid, connection)
			}

			if err != nil {
				logger.Error(err)
			}

			logger.Infof("delete application: %s", uid)

			return nil
		},
		Before: before,
	},
	{
		Name:      "app:list",
		Category:  category,
		Usage:     "list all applications",
		UsageText: "app:list [command options]",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "page",
				Aliases: []string{"p"},
				Usage:   "page code",
				Value:   1,
			},
			&cli.UintFlag{
				Name:    "size",
				Aliases: []string{"s"},
				Usage:   "size per page",
				Value:   15,
			},
			&cli.BoolFlag{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "with deleted",
				Value:   false,
			},
		},
		Before: before,
		Action: func(ctx *cli.Context) error {
			var (
				page    = ctx.Uint("page")
				size    = ctx.Uint("size")
				del     = ctx.Bool("delete")
				headers = []string{"ID", "UID", "Secret", "Name", "Note", "CreatedAt"}
				data    []string
				apps    []models.App
			)
			if page < 1 || size < 1 {
				return errors.New("page and size must be greater than 0")
			}
			if del {
				connection = connection.Unscoped()
				headers = append(headers, "DeletedAt")
			}
			connection.Offset((page - 1) * size).Limit(size).Find(&apps)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(headers)
			for _, app := range apps {
				data = []string{
					strconv.FormatUint(app.ID, 10),
					app.UID,
					app.Secret,
					app.Name,
					*app.Note,
					app.CreatedAt.Format("2006-01-02 15:04:05"),
				}
				if del {
					if app.DeletedAt != nil {
						data = append(data, app.DeletedAt.Format("2006-01-02 15:04:05"))
					} else {
						data = append(data, "")
					}
				}
				table.Append(data)
			}
			table.Render()
			return nil
		},
	},
}
