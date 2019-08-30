//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bigfile/bigfile/rpc"
	"google.golang.org/grpc"
)

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

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	tokenDelete(conn)
}
