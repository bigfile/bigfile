//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

// Package rpc provide rpc service

/*
Before calling rpc service, you should create a rpc connection, let's
see a complete example.

TokenCreate

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


TokenUpdate

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

TokenDelete

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

FileCreate

file upload may be a focus of attention, so let's see a complete Example:

	package main

	import (
		"context"
		"fmt"
		"io"
		"log"
		"os"

		"github.com/bigfile/bigfile/databases/models"
		"github.com/bigfile/bigfile/rpc"
		"github.com/golang/protobuf/ptypes/wrappers"
		"google.golang.org/grpc"
	)

	func uploadFile(conn *grpc.ClientConn) {
		var (
			ctx            context.Context
			err            error
			resp           *rpc.FileCreateResponse
			cancel         context.CancelFunc
			client         rpc.FileCreateClient
			streamClient   rpc.FileCreate_FileCreateClient
			waitUploadFile *os.File
		)
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		client = rpc.NewFileCreateClient(conn)
		if streamClient, err = client.FileCreate(ctx); err != nil {
			fmt.Println(err)
			return
		}
		if waitUploadFile, err = os.Open("/Users/fudenglong/Downloads/11.mp4"); err != nil {
			fmt.Println(err)
			return
		}
		defer waitUploadFile.Close()
		for index := 0; ; index++ {
			var chunk = make([]byte, models.ChunkSize*2)
			var size int
			var quit bool
			if size, err = waitUploadFile.Read(chunk); err != nil {
				if err != io.EOF {
					fmt.Println(err)
					return
				}
				quit = true
			}
			req := &rpc.FileCreateRequest{
				Token:   "bf0776c565412060eb93f8f307fae299",
				Path:    "/Users/fudenglong/Downloads/shield_agents.mp4",
				Content: &wrappers.BytesValue{Value: chunk[:size]},
			}
			if index == 0 {
				req.Operation = &rpc.FileCreateRequest_Rename{Rename: true}
			} else {
				req.Operation = &rpc.FileCreateRequest_Append{Append: true}
			}
			fmt.Println("sending request")
			if err = streamClient.Send(req); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("waiting resp")
			if resp, err = streamClient.Recv(); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(resp)
			if quit {
				break
			}
		}
	}

	func main() {
		// Set up a connection to the server.
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		uploadFile(conn)
	}

DirectoryCreate

Actually, in bigfile, file and directory both are regarded as 'file'. So, you can create a directory, Example:

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

FileUpdate

move file or directory and hide file, Example:

	func fileUpdate(conn *grpc.ClientConn) {
		c := rpc.NewFileUpdateClient(conn)
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		resp, err := c.FileUpdate(ctx, &rpc.FileUpdateRequest{
			Token:   "bf0776c565412060eb93f8f307fae299",
			FileUid: "556e3b9c936202c9dc67b7ad45530790",
			Path:    "/new/path/to/shield_agents.mp4",
			Hidden:  &wrappers.BoolValue{Value: true},
		})
		fmt.Println(resp, err)
	}

FileDelete

delete a file or a directory, Example:

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
*/

package rpc
