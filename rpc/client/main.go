//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bigfile/bigfile/rpc"
	"google.golang.org/grpc"
)

func createDir(conn *grpc.ClientConn) {
	var (
		ctx          context.Context
		err          error
		resp         *rpc.FileCreateResponse
		cancel       context.CancelFunc
		client       rpc.FileCreateClient
		streamClient rpc.FileCreate_FileCreateClient
	)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client = rpc.NewFileCreateClient(conn)
	if streamClient, err = client.FileCreate(ctx); err != nil {
		fmt.Println(err)
		return
	}
	req := &rpc.FileCreateRequest{
		Token:     "bf0776c565412060eb93f8f307fae299",
		Path:      "/create/some/directories",
		Operation: &rpc.FileCreateRequest_CreateDir{CreateDir: true},
	}
	if err = streamClient.Send(req); err != nil {
		fmt.Println(err)
		return
	}
	if resp, err = streamClient.Recv(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	createDir(conn)
}
