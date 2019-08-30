//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

/* Package rpc provide rpc service, It provides the following service:

Before calling rpc service, you should create a rpc connection, let's
see a complete example:

TokenCreate:

	package main

	import (
		"context"
		"fmt"
		"log"
		"time"

		"github.com/bigfile/bigfile/databases/models"
		"github.com/bigfile/bigfile/rpc"
		"github.com/golang/protobuf/ptypes/timestamp"
		"github.com/golang/protobuf/ptypes/wrappers"
		"google.golang.org/grpc"
	)

	func tokenCreate(conn *grpc.ClientConn) {
		c := rpc.NewTokenCreateClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.TokenCreate(ctx, &rpc.TokenCreateRequest{
			AppUid:         "1a951487fb16798c0c6d838decfbc973",
			AppSecret:      "38c57333fe2e2c17cc663f61212d7b7e",
			Path:           &wrappers.StringValue{Value: "/save/to/images"},
			Ip:             &wrappers.StringValue{Value: "192.168.0.0.1"},
			Secret:         &wrappers.StringValue{Value: models.RandomWithMd5(22)},
			AvailableTimes: &wrappers.UInt32Value{Value: 1000},
			ReadOnly:       &wrappers.BoolValue{Value: true},
			ExpiredAt:      &timestamp.Timestamp{Seconds: time.Now().Add(10 * time.Minute).Unix()},
		})
		fmt.Println(r, err)
	}


	func main() {
		// Set up a connection to the server.
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		tokenCreate(conn)
	}


TokenUpdate:

	func tokenUpdate(conn *grpc.ClientConn) {
		c := rpc.NewTokenUpdateClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.TokenUpdate(ctx, &rpc.TokenUpdateRequest{
			AppUid:         "1a951487fb16798c0c6d838decfbc973",
			AppSecret:      "38c57333fe2e2c17cc663f61212d7b7e",
			Token:          "bd5216fa7a6b5c5fdc8a250bae52b306",
			Path:           &wrappers.StringValue{Value: "/new/path"},
			Ip:             &wrappers.StringValue{Value: "192.168.0.2"},
			Secret:         &wrappers.StringValue{Value: models.RandomWithMd5(233)},
			AvailableTimes: &wrappers.UInt32Value{Value: 223},
			ReadOnly:       &wrappers.BoolValue{Value: true},
			ExpiredAt:      &timestamp.Timestamp{Seconds: time.Now().Add(1000 * time.Minute).Unix()},
		})
		fmt.Println(r, err)
	}

TokenDelete:

	func tokenDelete(conn *grpc.ClientConn) {
		c := rpc.NewTokenDeleteClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.TokenDelete(ctx, &rpc.TokenDeleteRequest{
			AppUid:    "1a951487fb16798c0c6d838decfbc973",
			AppSecret: "38c57333fe2e2c17cc663f61212d7b7e",
			Token:     "bd5216fa7a6b5c5fdc8a250bae52b306",
		})
		fmt.Println(r, err)
	}
*/

package rpc
