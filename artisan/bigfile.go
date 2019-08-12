//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	cmdApp "github.com/bigfile/bigfile/artisan/app"
	"github.com/bigfile/bigfile/artisan/http"
	"github.com/bigfile/bigfile/artisan/migrate"
	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/urfave/cli.v2"
)

var (
	app = cli.App{
		Name:     "bigfile",
		Version:  "0.1.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  "bigfile team",
				Email: "bigfile@gmail.com",
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
			if err = viper.ReadInConfig(); err != nil {
				return err
			}
			if err = viper.Unmarshal(config.DefaultConfig); err != nil {
				return err
			}
			if _, err = log.NewLogger(&config.DefaultConfig.Log); err != nil {
				return err
			}
			return nil
		},
	}
)

func main() {
	var commands []*cli.Command

	// just for develop environment
	if mode, ok := os.LookupEnv("BIGFILE_MODE"); ok && strings.HasPrefix(strings.ToLower(mode), "dev") {
		commands = append(commands, migrate.Commands...)
	}

	commands = append(commands, cmdApp.Commands...)
	commands = append(commands, http.Commands...)
	app.Commands = commands

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
