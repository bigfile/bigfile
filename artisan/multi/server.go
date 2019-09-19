//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package multi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	libHTTP "net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/bigfile/bigfile/ftp"
	"github.com/bigfile/bigfile/http"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/bigfile/bigfile/log"
	"github.com/bigfile/bigfile/rpc"
	"github.com/gin-gonic/gin"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/op/go-logging"
	"goftp.io/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
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
	logger = log.MustNewLogger(nil)

	category = "multi"

	// Commands is used to start multiple service simultaneously
	Commands = []*cli.Command{
		{
			Name:      "multi:server",
			Category:  category,
			Usage:     "start multiple service simultaneously",
			UsageText: "multi:server [command options]",
			Action: func(ctx *cli.Context) error {
				var (
					wg  sync.WaitGroup
					sig = make(chan struct{})
				)

				if !util.IsFile(ctx.String("server-cert")) {
					logger.Error("server certificate file not exist")
					return nil
				}

				if !util.IsFile(ctx.String("server-key")) {
					logger.Error("server certificate key file not exist")
					return nil
				}

				if !util.IsFile(ctx.String("ca-cert")) {
					logger.Error("ca certificate file not exist")
					return nil
				}

				wg.Add(3)

				go func() {
					defer wg.Done()
					_ = startHTTPServer(ctx, sig)
				}()

				go func() {
					defer wg.Done()
					_ = startFtpServer(ctx, sig)
				}()

				go func() {
					defer wg.Done()
					_ = startRPCServer(ctx, sig)
				}()

				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				close(sig)
				wg.Wait()
				return nil
			},
			Flags: []cli.Flag{
				// common parameters
				&cli.StringFlag{
					Name:    "host",
					Aliases: []string{"H"},
					Usage:   "http service listen ip",
					Value:   "0.0.0.0",
				},
				&cli.StringFlag{
					Name:  "server-cert",
					Usage: "certificate file for enable tls/ssl server",
					Value: "server.pem",
				},
				&cli.StringFlag{
					Name:  "server-key",
					Usage: "certificate key file for enable tls/ssl server",
					Value: "server.key",
				},
				&cli.StringFlag{
					Name:  "ca-cert",
					Usage: "certificate key file for enable tls/ssl server",
					Value: "ca.pem",
				},
				// http parameters
				&cli.Int64Flag{
					Name:  "http-port",
					Usage: "http service listen port",
					Value: 10985,
				},
				&cli.DurationFlag{
					Name: "http-read-timeout",
					Usage: "http-read-timeout is the maximum duration for reading " +
						"the entire request, including the body",
					Value: 0,
				},
				&cli.DurationFlag{
					Name: "http-read-header-timeout",
					Usage: "http-read-header-timeout is the amount of time allowed " +
						"to read request headers",
					Value: 0,
				},
				&cli.DurationFlag{
					Name: "http-write-timeout",
					Usage: "http-writer-timeout is the maximum duration before timing " +
						"out writes of the response",
					Value: 0,
				},
				&cli.DurationFlag{
					Name: "http-idle-timeout",
					Usage: "http-idle-timeout is the maximum amount of time to wait for " +
						"the next request when keep-alives are enabled",
					Value: 0,
				},
				&cli.DurationFlag{
					Name:  "http-wait-shutdown",
					Usage: "wait time before timeout for closing server",
					Value: 5 * time.Second,
				},
				&cli.IntFlag{
					Name: "http-max-header-bytes",
					Usage: "http-max-header-bytes controls the maximum number of bytes the " +
						"server will read parsing the request header's keys and values, " +
						"including the request line. It does not limit the size of the request body",
					Value: 0,
				},
				// ftp parameters
				&cli.UintFlag{
					Name:  "ftp-port",
					Usage: "ftp server port",
					Value: 2121,
				},
				&cli.StringFlag{
					Name:  "ftp-passive-ip",
					Usage: "ftp client connect to ftp server to transfer data in passive mode",
					Value: "",
				},
				&cli.StringFlag{
					Name:  "ftp-passive-port-range",
					Usage: "ftp server will pick a port random in the range, open the data connection tunnel in passive mode",
					Value: "52013-52114",
				},
				&cli.StringFlag{
					Name:  "ftp-welcome-message",
					Usage: "ftp server welcome message",
					Value: "welcome to bigfile ftp server",
				},
				// rpc parameter
				&cli.IntFlag{
					Name:  "rpc-auth-client",
					Usage: "rpc service client certificate auth type, 0: no client cert, 1: request client cert, 2: require any client cert, 3: verify client cert if given, 4: require and verify client cert",
					Value: 4,
				},
				&cli.Int64Flag{
					Name:  "rpc-port",
					Usage: "rpc service listen port",
					Value: 10986,
				},
			},
			Before: func(ctx *cli.Context) (err error) {
				gin.SetMode(gin.ReleaseMode)
				db := databases.MustNewConnection(&config.DefaultConfig.Database)
				migrate.DefaultMC.SetConnection(db)
				migrate.DefaultMC.Upgrade()
				return nil
			},
		},
	}
)

func startHTTPServer(ctx *cli.Context, sig chan struct{}) error {
	addr := fmt.Sprintf("%s:%d", ctx.String("host"), ctx.Int64("http-port"))
	s := libHTTP.Server{
		Addr:              addr,
		Handler:           http.Routers(),
		ReadTimeout:       ctx.Duration("http-read-timeout"),
		ReadHeaderTimeout: ctx.Duration("http-read-header-timeout"),
		WriteTimeout:      ctx.Duration("http-write-timeout"),
		IdleTimeout:       ctx.Duration("http-idle-timeout"),
		MaxHeaderBytes:    ctx.Int("http-max-header-bytes"),
	}
	go func() {
		logger.Debugf("bigfile http service listening on: https://%s", addr)
		if err := s.ListenAndServeTLS(ctx.String("server-cert"), ctx.String("server-key")); err != nil && err != libHTTP.ErrServerClosed {
			logger.Errorf("https server error: %s", err)
		}
	}()
	<-sig
	logger.Debug("Shutdown HTTP Server ...")
	c, cancel := context.WithTimeout(context.Background(), ctx.Duration("http-wait-shutdown"))
	defer cancel()
	if err := s.Shutdown(c); err != nil {
		logger.Fatal("HTTP Server Shutdown:", err)
	}
	<-c.Done()
	logger.Debugf("HTTP Shutdown timeout of %s", ctx.Duration("http-wait-shutdown"))
	logger.Debug("HTTP Server exiting")
	return nil
}

func startFtpServer(ctx *cli.Context, sig chan struct{}) error {
	logger := &loggerAdapter{logger}
	passivePortRange := ctx.String("ftp-passive-port-range")
	if len(passivePortRange) > 0 {
		portRange := strings.Split(passivePortRange, "-")
		if len(portRange) != 2 {
			return errors.New("wrong ftp-passive-port-range format, eg: 52013-52114")
		}
		if _, err := strconv.Atoi(portRange[0]); err != nil {
			return err
		}
		if _, err := strconv.Atoi(portRange[1]); err != nil {
			return err
		}
		if portRange[0] >= portRange[1] {
			return errors.New("ftp port start should be less than port end")
		}
	}
	host := ctx.String("host")
	port := int(ctx.Uint("ftp-port"))
	options := &server.ServerOpts{
		TLS:            true,
		Auth:           &ftp.Auth{},
		Port:           port,
		Logger:         logger,
		Factory:        &ftp.Factory{},
		KeyFile:        ctx.String("server-key"),
		Hostname:       host,
		CertFile:       ctx.String("server-cert"),
		PublicIp:       ctx.String("ftp-passive-ip"),
		PassivePorts:   passivePortRange,
		ExplicitFTPS:   true,
		WelcomeMessage: ctx.String("ftp-welcome-message"),
	}
	go func() {
		if err := server.NewServer(options).ListenAndServe(); err != nil {
			logger.log.Errorf("ftp server start with error: %s", err)
		}
	}()
	<-sig
	log.MustNewLogger(nil).Debug("Shutdown FTP Server ...")
	return nil
}

func startRPCServer(ctx *cli.Context, sig chan struct{}) (err error) {
	var (
		serverCert         tls.Certificate
		certPool           = x509.NewCertPool()
		rootCaContentBytes []byte
		tlsConf            *tls.Config
		listener           net.Listener
	)
	if serverCert, err = tls.LoadX509KeyPair(ctx.String("server-cert"), ctx.String("server-key")); err != nil {
		log.MustNewLogger(nil).Errorf("rpc, load server certificates failed, %s", err)
		return
	}
	if rootCaContentBytes, err = ioutil.ReadFile(ctx.String("ca-cert")); err != nil {
		log.MustNewLogger(nil).Errorf("rpc, load ca certificates failed, %s", err)
		return
	}
	if !certPool.AppendCertsFromPEM(rootCaContentBytes) {
		return errors.New("rpc, append root ca to cert pool failed")
	}
	tlsConf = &tls.Config{
		ClientAuth:   tls.ClientAuthType(ctx.Int("rpc-auth-client")),
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
	}
	addr := fmt.Sprintf("%s:%d", ctx.String("host"), ctx.Int64("rpc-port"))
	if listener, err = net.Listen("tcp", addr); err != nil {
		return err
	}
	defer listener.Close()
	rpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConf)),
		grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: 2 * time.Minute}),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_prometheus.StreamServerInterceptor,
				grpc_recovery.StreamServerInterceptor(),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				grpc_recovery.UnaryServerInterceptor(),
			),
		),
	)

	service := &rpc.Server{}
	rpc.RegisterDirectoryListServer(rpcServer, service)
	rpc.RegisterTokenCreateServer(rpcServer, service)
	rpc.RegisterTokenUpdateServer(rpcServer, service)
	rpc.RegisterTokenDeleteServer(rpcServer, service)
	rpc.RegisterFileCreateServer(rpcServer, service)
	rpc.RegisterFileReadServer(rpcServer, service)
	rpc.RegisterFileUpdateServer(rpcServer, service)
	rpc.RegisterFileDeleteServer(rpcServer, service)

	go func() {
		log.MustNewLogger(nil).Debugf("bigfile rpc service listening on: tcp://%s", listener.Addr().String())
		if err = rpcServer.Serve(listener); err != nil {
			log.MustNewLogger(nil).Error(err)
		}
	}()
	<-sig
	log.MustNewLogger(nil).Debug("Shutdown RPC Server ...")
	rpcServer.GracefulStop()
	return nil
}
