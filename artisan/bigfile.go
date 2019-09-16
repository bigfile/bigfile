//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package main is the entry for the whole program
package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	cmdApp "github.com/bigfile/bigfile/artisan/app"
	"github.com/bigfile/bigfile/artisan/ftp"
	"github.com/bigfile/bigfile/artisan/http"
	"github.com/bigfile/bigfile/artisan/migrate"
	"github.com/bigfile/bigfile/artisan/rpc"
	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/log"
	"github.com/gookit/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/urfave/cli.v2"
)

var (
	app = cli.App{
		Name:     "bigfile",
		Version:  "1.0.2",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "bigfile team",
				Email: "bigfilefu@gmail.com",
			},
		},
		Copyright: "copyright 2019 The bigfile authors",
		Usage:     "develop toolkit and program entry",
		UsageText: "bigfile [global options] command [command options] [arguments...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "system config file，search path: .bigfile.yaml, $HOME/.bigfile.yaml, /etc/bigfile/.bigfile.yaml",
				Aliases: []string{"c"},
				EnvVars: []string{"BIGFILE_CONFIG"},
			},
			&cli.StringFlag{
				Name:  "db-host",
				Usage: "set the database host",
				Value: "127.0.0.1",
			},
			&cli.UintFlag{
				Name:  "db-port",
				Usage: "set the database port",
				Value: 3306,
			},
			&cli.StringFlag{
				Name:  "db-user",
				Usage: "set the database user",
				Value: "root",
			},
			&cli.StringFlag{
				Name:  "db-pass",
				Usage: "set the database password",
				Value: "root",
			},
			&cli.StringFlag{
				Name:  "db-name",
				Usage: "set the database name",
				Value: "bigfile",
			},
		},
		Before: func(ctx *cli.Context) error {
			var (
				err      error
				cfgFile  = ctx.String("config")
				userHome string
			)
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
			} else {
				viper.SetConfigName(".bigfile")
				viper.AddConfigPath(".")
				if userHome, err = homedir.Dir(); err != nil {
					return err
				}
				viper.AddConfigPath(userHome)
				viper.AddConfigPath("/etc/bigfile")
			}

			if err = viper.ReadInConfig(); err == nil {
				if err = viper.Unmarshal(config.DefaultConfig); err != nil {
					return err
				}
				if _, err = log.NewLogger(&config.DefaultConfig.Log); err != nil {
					return err
				}
			} else {
				color.Warn.Println(err.Error())
				config.DefaultConfig.Database.Host = ctx.String("db-host")
				config.DefaultConfig.Database.Port = uint32(ctx.Uint("db-port"))
				config.DefaultConfig.Database.User = ctx.String("db-user")
				config.DefaultConfig.Database.Password = ctx.String("db-pass")
				config.DefaultConfig.Database.DBName = ctx.String("db-name")
			}

			return nil
		},
	}
)

func main() {
	var commands []*cli.Command

	commands = append(commands, migrate.Commands...)
	commands = append(commands, cmdApp.Commands...)
	commands = append(commands, http.Commands...)
	commands = append(commands, rpc.Commands...)
	commands = append(commands, ftp.Commands...)
	app.Commands = commands

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
