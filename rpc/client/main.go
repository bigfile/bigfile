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

func fileDelete(conn *grpc.ClientConn) {
	c := rpc.NewFileDeleteClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.FileDelete(ctx, &rpc.FileDeleteRequest{
		Token:            "bf0776c565412060eb93f8f307fae299",
		FileUid:          "556e3b9c936202c9dc67b7ad45530790",
		ForceDeleteIfDir: false,
	})
	fmt.Println(resp, err)
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	fileDelete(conn)
}
