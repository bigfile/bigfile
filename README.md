# Bigfile —— Manage Files In A Different Way

<img align="right" width="159px" src="https://avatars3.githubusercontent.com/u/52916753">

[![Build Status](https://travis-ci.org/bigfile/bigfile.svg?branch=master)](https://travis-ci.org/bigfile/bigfile)
[![codecov](https://codecov.io/gh/bigfile/bigfile/branch/master/graph/badge.svg)](https://codecov.io/gh/bigfile/bigfile)
[![GoDoc](https://godoc.org/github.com/bigfile/bigfile?status.svg)](https://github.com/bigfile/bigfile)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/bigfile/bigfile)](https://goreportcard.com/report/github.com/bigfile/bigfile)
[![Open Source Helpers](https://www.codetriage.com/bigfile/bigfile/badges/users.svg)](https://www.codetriage.com/bigfile/bigfile)

Bigfile is built on top of many excellent open source projects. Designed to provide a file management service and give developers more help. At the bottom, bigfile splits the file into small pieces of 1MB, only the last shard may be less than 1mb, the same slice will only be stored once.

We also built a virtual file organization system that is logically divided into directories and files, files and directories can be deleted, moved, added and updated. You can also use the ftp service we provide to manage files. That's just too cool. In the development project, you only need to use the http api to access, in the future we will also develop various sdk, reduce the difficulty of use.


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_large)



### RPC Protocol

#### generate certificates

>  go run artisan/bigfile.go rpc create-cert

#### start rpc service

> go run artisan/bigfile.go rpc start --server-cert server.pem --server-key server.key --root-cert root.pem --auth-client 4

#### connect to remote server

```go
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/bigfile/bigfile/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	var (
		err           error
		conn          *grpc.ClientConn
		cert          tls.Certificate
		certPool      *x509.CertPool
		rootCertBytes []byte
	)
	if cert, err = tls.LoadX509KeyPair("client.pem", "client.key"); err != nil {
		fmt.Println(err)
		return
	}

	certPool = x509.NewCertPool()
	if rootCertBytes, err = ioutil.ReadFile("ca.pem"); err != nil {
		fmt.Println(err)
		return
	}

	if !certPool.AppendCertsFromPEM(rootCertBytes) {
		fmt.Print("Fail to append ca")
		return
	}

	if conn, err = grpc.Dial("127.0.0.1:10986", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}))); err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	grpcClient := rpc.NewTokenCreateClient(conn)
	fmt.Println(grpcClient.TokenCreate(context.TODO(), &rpc.TokenCreateRequest{
		AppUid:    "1a951487fb16798c0c6d838decfbc973",
		AppSecret: "38c57333fe2e2c17cc663f61212d7b7e",
	}))
}
```