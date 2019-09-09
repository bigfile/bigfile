//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package ftp is the entry for tfp service
package ftp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/bigfile/bigfile/ftp"
	"github.com/bigfile/bigfile/log"
	"github.com/op/go-logging"
	"goftp.io/server"
	"gopkg.in/urfave/cli.v2"

	// import migration
	_ "github.com/bigfile/bigfile/databases/migrate/migrations"
)

type loggerAdapter struct {
	log *logging.Logger
}

func (l *loggerAdapter) Print(sessionID string, message interface{}) {
	l.log.Debug(sessionID, message)
}
func (l *loggerAdapter) Printf(sessionID string, format string, v ...interface{}) {
	l.log.Debug(sessionID, fmt.Sprintf(format, v...))
}
func (l *loggerAdapter) PrintCommand(sessionID string, command string, params string) {
	if command == "PASS" {
		l.log.Debugf("%s > PASS ****", sessionID)
	} else {
		l.log.Debugf("%s > %s %s", sessionID, command, params)
	}
}
func (l *loggerAdapter) PrintResponse(sessionID string, code int, message string) {
	l.log.Debugf("%s < %d %s", sessionID, code, message)
}

var (
	category = "ftp"

	// Commands represent the ftp service start command
	Commands = []*cli.Command{
		{
			Name:      "ftp:start",
			Category:  category,
			Usage:     "start ftp service",
			UsageText: "ftp:start [command options]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "host",
					Aliases: []string{"H"},
					Usage:   "ftp server host",
					Value:   "::",
				},
				&cli.UintFlag{
					Name:    "port",
					Aliases: []string{"P"},
					Usage:   "ftp server port",
					Value:   2121,
				},
				&cli.StringFlag{
					Name:  "passive-ip",
					Usage: "ftp client connect to ftp server to transfer data in passive mode",
					Value: "",
				},
				&cli.StringFlag{
					Name:  "passive-port-range",
					Usage: "ftp server will pick a port random in the range, open the data connection tunnel in passive mode",
					Value: "52013-52114",
				},
				&cli.StringFlag{
					Name:  "welcome-message",
					Usage: "ftp server welcome message",
					Value: "welcome to bigfile ftp server",
				},
				&cli.BoolFlag{
					Name:  "tls-enable",
					Usage: "determine whether enable tls",
					Value: false,
				},
				&cli.StringFlag{
					Name:  "cert-file",
					Usage: "tls certificate file",
					Value: "",
				},
				&cli.StringFlag{
					Name:  "key-file",
					Usage: "tls certificate key file",
					Value: "",
				},
			},
			Action: func(ctx *cli.Context) error {
				logger := &loggerAdapter{log.MustNewLogger(nil)}
				passivePortRange := ctx.String("passive-port-range")
				if len(passivePortRange) > 0 {
					portRange := strings.Split(passivePortRange, "-")
					if len(portRange) != 2 {
						return errors.New("wrong passive-port-range format, eg: 52013-52114")
					}
					if _, err := strconv.Atoi(portRange[0]); err != nil {
						return err
					}
					if _, err := strconv.Atoi(portRange[1]); err != nil {
						return err
					}
					if portRange[0] >= portRange[1] {
						return errors.New("port start should be less than port end")
					}
				}
				host := ctx.String("host")
				port := int(ctx.Uint("port"))
				options := &server.ServerOpts{
					TLS:            ctx.Bool("tls-enable"),
					Auth:           &ftp.Auth{},
					Port:           port,
					Logger:         logger,
					Factory:        &ftp.Factory{},
					KeyFile:        ctx.String("key-file"),
					Hostname:       host,
					CertFile:       ctx.String("cert-file"),
					PublicIp:       ctx.String("passive-ip"),
					PassivePorts:   passivePortRange,
					ExplicitFTPS:   ctx.Bool("tls-enable"),
					WelcomeMessage: ctx.String("welcome-message"),
				}
				return server.NewServer(options).ListenAndServe()
			},
			Before: func(context *cli.Context) (err error) {
				db := databases.MustNewConnection(&config.DefaultConfig.Database)
				if err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.DefaultConfig.Database.DBName)).Error; err != nil {
					return
				}
				migrate.DefaultMC.SetConnection(db)
				migrate.DefaultMC.Upgrade()
				return nil
			},
		},
	}
)
