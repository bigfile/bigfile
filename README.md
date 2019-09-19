### Bigfile ———— A Powerful File Transfer System with "逼格" [简体中文](https://learnku.com/docs/bigfile) ∙ [English](https://bigfile.site)
----

<p align="center">
    <a href="https://travis-ci.org/bigfile/bigfile"><img src="https://travis-ci.org/bigfile/bigfile.svg?branch=master"/></a>
    <a href="https://codecov.io/gh/bigfile/bigfile"><img src="https://codecov.io/gh/bigfile/bigfile/branch/master/graph/badge.svg"/></a>
    <a href="https://godoc.org/github.com/bigfile/bigfile"><img src="https://godoc.org/github.com/bigfile/bigfile?status.svg"/></a>
    <a href="https://github.com/bigfile/bigfile"><img src="https://img.shields.io/badge/release-v1.0.8-blue"/></a>
    <a href="https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_shield"><img src="https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=shield"/></a>
    <a href="https://goreportcard.com/report/github.com/bigfile/bigfile"><img src="https://goreportcard.com/badge/github.com/bigfile/bigfile"/></a>
    <a href="https://golangci.com/r/github.com/bigfile/bigfile"><img src="https://golangci.com/badges/github.com/bigfile/bigfile.svg"/></a>
    <a href="https://www.codetriage.com/bigfile/bigfile"><img src="https://www.codetriage.com/bigfile/bigfile/badges/users.svg"/></a>
</p>

<p align="center">
    <img src="https://bigfile.site/bigfile.png" />
</p>

**Bigfile** is a file transfer system, supports http, ftp and rpc protocol. Designed to provide a file management service and give developers more help. At the bottom, bigfile splits the file into small pieces of **1MB**, the same slice will only be stored once.

In fact, we built a file organization system based on the database. Here you can find familiar files and folders.

Since the rpc and http protocols are supported, those languages supported by [grpc](https://grpc.io/) and other languages can be quickly accessed.

More detailed documents can be found here

### Features

* Support HTTP(s) protocol

    * Support rate limit by ip
    * Support cors
    * Support to avoid replay attack
    * Support to validate parameter signature
    * Support Http Single Range Protocol

* Support FTP(s) protocol

* Support RPC protocol

* Support to deploy by [docker](https://hub.docker.com/r/bigfile/bigfile)

* Provide document with English and Chinese


### Install Bigfile

There are kinds of ways to install Bigfile, you can find more detailed documentation here [English](https://bigfile.site/start/) [简体中文](https://learnku.com/docs/bigfile)

### Start Bigfile

1. generate certificates

> bigfile rpc:make-cert

2. start server

>  bigfile multi:server
>

This will print some information as follows:

        [2019/09/19 15:38:32.817] 56628 DEBUG  bigfile http service listening on: https://0.0.0.0:10985
        [2019/09/19 15:38:32.818] 56628 DEBUG   Go FTP Server listening on 2121
        [2019/09/19 15:38:32.819] 56628 DEBUG  bigfile rpc service listening on: tcp://[::]:10986


### HTTP Example (Token Create)

```go
package main

import (
	"fmt"
	"io/ioutil"
	libHttp "net/http"
	"strings"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/http"
)

func main() {
	appUid := "42c4fcc1a620c9e97188f50b6f2ab199"
	appSecret := "f8f2ae1fe4f70b788254dcc991a35558"
	body := http.GetParamsSignBody(map[string]interface{}{
		"appUid":         appUid,
		"nonce":          models.RandomWithMd5(128),
		"path":           "/images/png",
		"expiredAt":      time.Now().AddDate(0, 0, 2).Unix(),
		"secret":         models.RandomWithMd5(44),
		"availableTimes": -1,
		"readOnly":       false,
	}, appSecret)
	request, err := libHttp.NewRequest(
		"POST", "https://127.0.0.1:10985/api/bigfile/token/create", strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := libHttp.DefaultClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(string(bodyBytes))
	}
}
```

### RPC Example (Token Create)

```go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc"
	"github.com/bigfile/bigfile/rpc"
	"google.golang.org/grpc/credentials"
)

func createConnection() (*grpc.ClientConn, error) {
	var (
		err           error
		conn          *grpc.ClientConn
		cert          tls.Certificate
		certPool      *x509.CertPool
		rootCertBytes []byte
	)
	if cert, err = tls.LoadX509KeyPair("client.pem", "client.key"); err != nil {
		return nil, err
	}

	certPool = x509.NewCertPool()
	if rootCertBytes, err = ioutil.ReadFile("ca.pem"); err != nil {
		return nil, err
	}

	if !certPool.AppendCertsFromPEM(rootCertBytes) {
		return nil, err
	}

	if conn, err = grpc.Dial("192.168.0.103:10986", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}))); err != nil {
		return nil, err
	}
	return conn, err
}

func main() {
	var (
		err  error
		conn *grpc.ClientConn
	)

	if conn, err = createConnection(); err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	grpcClient := rpc.NewTokenCreateClient(conn)
    	fmt.Println(grpcClient.TokenCreate(context.TODO(), &rpc.TokenCreateRequest{
    		AppUid:    "42c4fcc1a620c9e97188f50b6f2ab199",
    		AppSecret: "f8f2ae1fe4f70b788254dcc991a35558",
    }))
}
```

### FTP Example

![ftp-example](https://bigfile.site/ftp-connect-succesfully.png)

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbigfile%2Fbigfile?ref=badge_large)
