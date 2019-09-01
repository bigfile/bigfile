//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

// Package rpc provide the rpc service entry and other tools
package rpc

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/migrate"
	"github.com/bigfile/bigfile/log"
	"github.com/bigfile/bigfile/rpc"
	"github.com/gookit/color"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/urfave/cli.v2"

	// 导入migration
	_ "github.com/bigfile/bigfile/databases/migrate/migrations"
)

var (
	category = "rpc"

	// Commands defines the rpc service command line entry
	Commands = []*cli.Command{
		{
			Name:      "rpc",
			Category:  category,
			Usage:     "run rpc service",
			UsageText: "rpc command [command options]",
			Subcommands: []*cli.Command{
				{
					Name:      "create-cert",
					Usage:     "generate root, client and server certificates",
					UsageText: "rpc cert [command options]",
					Action:    generateCertificates,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "root-cert-out",
							Usage: "root ca certificate cert save path",
							Value: "root.pem",
						},
						&cli.StringFlag{
							Name:  "root-key-out",
							Usage: "root ca certificate key save path",
							Value: "root.key",
						},
						&cli.StringFlag{
							Name:  "server-cert-out",
							Usage: "server ca certificate cert save path",
							Value: "server.pem",
						},
						&cli.StringFlag{
							Name:  "server-key-out",
							Usage: "server ca certificate key save path",
							Value: "server.key",
						},
						&cli.StringFlag{
							Name:  "client-cert-out",
							Usage: "client ca certificate cert save path",
							Value: "client.pem",
						},
						&cli.StringFlag{
							Name:  "client-key-out",
							Usage: "client ca certificate key save path",
							Value: "client.key",
						},
					},
				},
				{
					Name:      "start",
					Usage:     "start rpc service",
					UsageText: "rpc start [command options]",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "server-cert",
							Usage: "the path to server certificate",
						},
						&cli.StringFlag{
							Name:  "server-key",
							Usage: "the path to server certificate key",
						},
						&cli.StringFlag{
							Name:  "root-cert",
							Usage: "the path to root certificate",
						},
						&cli.IntFlag{
							Name:  "auth-client",
							Usage: "client certificate auth type, 0: no client cert, 1: request client cert, 2: require any client cert, 3: verify client cert if given, 4: require and verify client cert",
							Value: 4,
						},
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
							Value:   10986,
						},
					},
					Action: func(ctx *cli.Context) (err error) {
						var (
							serverCert         tls.Certificate
							certPool           = x509.NewCertPool()
							rootCaContentBytes []byte
							tlsConf            *tls.Config
							listener           net.Listener
						)
						if serverCert, err = tls.LoadX509KeyPair(ctx.String("server-cert"), ctx.String("server-key")); err != nil {
							return
						}
						if rootCaContentBytes, err = ioutil.ReadFile(ctx.String("root-cert")); err != nil {
							return
						}
						if !certPool.AppendCertsFromPEM(rootCaContentBytes) {
							return errors.New("append root ca to cert pool failed")
						}
						tlsConf = &tls.Config{
							ClientAuth:   tls.ClientAuthType(ctx.Int("auth-client")),
							Certificates: []tls.Certificate{serverCert},
							ClientCAs:    certPool,
						}
						addr := fmt.Sprintf("%s:%d", ctx.String("host"), ctx.Int64("port"))
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
							log.MustNewLogger(nil).Infof("bigfile rpc service listening on: tcp://%s", listener.Addr().String())
							if err = rpcServer.Serve(listener); err != nil {
								log.MustNewLogger(nil).Error(err)
							}
						}()

						quit := make(chan os.Signal, 1)
						signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
						<-quit
						log.MustNewLogger(nil).Debug("Shutdown Server ...")
						rpcServer.GracefulStop()

						return
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
			},
		},
	}
)

func generateCertAndKey(template, parent *x509.Certificate, name, crtOut, keyOut string, parentPrivateKey *rsa.PrivateKey) (privateKey *rsa.PrivateKey, err error) {
	var (
		caBytes       []byte
		rootCaCrtFile *os.File
		rootCaKeyFile *os.File
	)
	if privateKey, err = rsa.GenerateKey(crand.Reader, 2048); err != nil {
		return
	}
	if parentPrivateKey == nil {
		parentPrivateKey = privateKey
	}
	if caBytes, err = x509.CreateCertificate(crand.Reader, template, parent, &privateKey.PublicKey, parentPrivateKey); err != nil {
		return
	}
	if rootCaCrtFile, err = os.Create(crtOut); err != nil {
		return
	}
	color.Green.Printf("generate %s certificate: %s\n", name, crtOut)
	defer rootCaCrtFile.Close()
	if err = pem.Encode(rootCaCrtFile, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return
	}
	color.Green.Printf("generate %s key: %s\n", name, keyOut)
	if rootCaKeyFile, err = os.Create(keyOut); err != nil {
		return
	}
	defer rootCaKeyFile.Close()
	if err = pem.Encode(rootCaKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return
	}
	return
}

func generateCertificates(ctx *cli.Context) (err error) {
	rootCa := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			Organization: []string{"Bigfile, INC"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	var rootPrivateKey *rsa.PrivateKey

	// generate root certificates
	if rootPrivateKey, err = generateCertAndKey(rootCa, rootCa, "root", ctx.String("root-cert-out"), ctx.String("root-key-out"), nil); err != nil {
		return
	}

	// generate server certificates
	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			Organization: []string{"Bigfile, INC"},
		},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv6loopback,
			net.IPv4(0, 0, 0, 0),
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	if _, err = generateCertAndKey(serverCert, rootCa, "server", ctx.String("server-cert-out"), ctx.String("server-key-out"), rootPrivateKey); err != nil {
		return
	}

	// generate client certificates
	clientCert := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			Organization: []string{"Bigfile, INC"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	if _, err = generateCertAndKey(clientCert, rootCa, "client", ctx.String("client-cert-out"), ctx.String("client-key-out"), rootPrivateKey); err != nil {
		return
	}

	return
}
