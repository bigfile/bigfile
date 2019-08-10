//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package http

import (
	ctx "context"
	"fmt"
	libHTTP "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bigfile/bigfile/http"
	"github.com/bigfile/bigfile/log"
	"github.com/gin-gonic/gin"
	"gopkg.in/urfave/cli.v2"
)

var (
	category = "http"
	logger   = log.MustNewLogger(nil)

	// Commands http service command
	Commands = []*cli.Command{
		{
			Name:      "http",
			Category:  category,
			Usage:     "run http service",
			UsageText: "http command [command options]",
			Subcommands: []*cli.Command{
				{
					Name:      "start",
					Usage:     "start http service",
					UsageText: "http start [command options]",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "host",
							Aliases: []string{"H"},
							Usage:   "http service listen ip",
							Value:   "0.0.0.0",
						},
						&cli.Int64Flag{
							Name:    "port",
							Aliases: []string{"P"},
							Usage:   "http service listen port",
							Value:   10985,
						},
						&cli.DurationFlag{
							Name: "read-timeout",
							Usage: "read-timeout is the maximum duration for reading " +
								"the entire request, including the body",
							Value: 0,
						},
						&cli.DurationFlag{
							Name: "read-header-timeout",
							Usage: "read-header-timeout is the amount of time allowed " +
								"to read request headers",
							Value: 0,
						},
						&cli.DurationFlag{
							Name: "write-timeout",
							Usage: "writer-timeout is the maximum duration before timing " +
								"out writes of the response",
							Value: 0,
						},
						&cli.DurationFlag{
							Name: "idle-timeout",
							Usage: "idle-timeout is the maximum amount of time to wait for " +
								"the next request when keep-alives are enabled",
							Value: 0,
						},
						&cli.DurationFlag{
							Name:  "wait-shutdown",
							Usage: "wait time before timeout for closing server",
							Value: 5 * time.Second,
						},
						&cli.IntFlag{
							Name: "max-header-bytes",
							Usage: "max-header-bytes controls the maximum number of bytes the " +
								"server will read parsing the request header's keys and values, " +
								"including the request line. It does not limit the size of the request body",
							Value: 0,
						},
						&cli.StringFlag{
							Name:  "cert-file",
							Usage: "certificate file for starting https service",
						},
						&cli.StringFlag{
							Name:  "cert-key",
							Usage: "certificate key file for starting https service",
						},
					},
					Action: func(context *cli.Context) error {
						gin.SetMode(gin.ReleaseMode)
						addr := fmt.Sprintf("%s:%d", context.String("host"), context.Int64("port"))
						server := libHTTP.Server{
							Addr:              addr,
							Handler:           http.Routers(),
							ReadTimeout:       context.Duration("read-timeout"),
							ReadHeaderTimeout: context.Duration("read-header-timeout"),
							WriteTimeout:      context.Duration("write-timeout"),
							IdleTimeout:       context.Duration("idle-timeout"),
							MaxHeaderBytes:    context.Int("max-header-bytes"),
						}
						certFile := context.String("cert-file")
						certKey := context.String("cert-key")

						go func() {
							if certFile != "" && certKey != "" {
								logger.Debugf("bigfile http service listening on: https://%s", addr)
								if err := server.ListenAndServeTLS(certFile, certKey); err != nil && err != libHTTP.ErrServerClosed {
									logger.Errorf("https server error: %s", err)
								}
							} else {
								logger.Debugf("bigfile http service listening on: http://%s", addr)
								if err := server.ListenAndServe(); err != nil && err != libHTTP.ErrServerClosed {
									logger.Errorf("https server error: %s", err)
								}

							}

						}()

						quit := make(chan os.Signal, 1)
						signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
						<-quit
						logger.Debug("Shutdown Server ...")

						ctx, cancel := ctx.WithTimeout(ctx.Background(), context.Duration("wait-shutdown"))
						defer cancel()
						if err := server.Shutdown(ctx); err != nil {
							logger.Fatal("Server Shutdown:", err)
						}
						<-ctx.Done()
						logger.Debugf("Shutdown timeout of %s", context.Duration("wait-shutdown"))
						logger.Debug("Server exiting")
						return nil
					},
				},
			},
		},
	}
)
