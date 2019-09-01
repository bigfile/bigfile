//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis = bufconn.Listen(bufSize)

func dialer() func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func init() {
	s := grpc.NewServer()
	server := &Server{}
	RegisterTokenCreateServer(s, server)
	RegisterTokenUpdateServer(s, server)
	RegisterTokenDeleteServer(s, server)
	RegisterFileCreateServer(s, server)
	RegisterFileReadServer(s, server)
	RegisterFileDeleteServer(s, server)
	RegisterFileUpdateServer(s, server)
	RegisterDirectoryListServer(s, server)
	go func() { _ = s.Serve(lis) }()
}

func getConn() *grpc.ClientConn {
	conn, _ := grpc.DialContext(newContext(context.Background()), "bufnet", grpc.WithContextDialer(dialer()), grpc.WithInsecure())
	return conn
}

func ExampleServer_TokenCreate() {
	c := NewTokenCreateClient(getConn())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.TokenCreate(ctx, &TokenCreateRequest{
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

func ExampleServer_TokenUpdate() {
	c := NewTokenUpdateClient(getConn())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.TokenUpdate(ctx, &TokenUpdateRequest{
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

func ExampleServer_TokenDelete() {
	c := NewTokenDeleteClient(getConn())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.TokenDelete(ctx, &TokenDeleteRequest{
		AppUid:    "1a951487fb16798c0c6d838decfbc973",
		AppSecret: "38c57333fe2e2c17cc663f61212d7b7e",
		Token:     "bd5216fa7a6b5c5fdc8a250bae52b306",
	})
	fmt.Println(r, err)
}

// ExampleServer_FileCreate is used to display how to upload a file
func ExampleServer_FileCreate() {
	var (
		ctx            context.Context
		err            error
		resp           *FileCreateResponse
		cancel         context.CancelFunc
		client         FileCreateClient
		streamClient   FileCreate_FileCreateClient
		waitUploadFile *os.File
	)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client = NewFileCreateClient(getConn())
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
		req := &FileCreateRequest{
			Token:   "bf0776c565412060eb93f8f307fae299",
			Path:    "/Users/fudenglong/Downloads/shield_agents.mp4",
			Content: &wrappers.BytesValue{Value: chunk[:size]},
		}
		if index == 0 {
			req.Operation = &FileCreateRequest_Rename{Rename: true}
		} else {
			req.Operation = &FileCreateRequest_Append{Append: true}
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

// ExampleServer_FileCreate2 is used to display how to create a directory
func ExampleServer_FileCreate2() {
	var (
		ctx          context.Context
		err          error
		resp         *FileCreateResponse
		cancel       context.CancelFunc
		client       FileCreateClient
		streamClient FileCreate_FileCreateClient
	)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client = NewFileCreateClient(getConn())
	if streamClient, err = client.FileCreate(ctx); err != nil {
		fmt.Println(err)
		return
	}
	req := &FileCreateRequest{
		Token:     "bf0776c565412060eb93f8f307fae299",
		Path:      "/create/some/directories",
		Operation: &FileCreateRequest_CreateDir{CreateDir: true},
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

func ExampleServer_FileRead() {
	var (
		err          error
		client       = NewFileReadClient(getConn())
		header       metadata.MD
		fileName     string
		fileHash     string
		dataHash     string
		dataBuffer   *bytes.Buffer
		fileSize     int
		streamClient FileRead_FileReadClient
	)
	if streamClient, err = client.FileRead(context.Background(), &FileReadRequest{
		Token:   "bf0776c565412060eb93f8f307fae299",
		FileUid: "556e3b9c936202c9dc67b7ad45530790",
	}); err != nil {
		fmt.Println(err)
		return
	}
	if header, err = streamClient.Header(); err != nil {
		fmt.Println(err)
		return
	}
	fileName = header.Get("name")[0]
	fileHash = header.Get("hash")[0]
	if fileSize, err = strconv.Atoi(header.Get("size")[0]); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("name = %s, hash = %s, size = %d\n", fileName, fileHash, fileSize)
	dataBuffer = new(bytes.Buffer)
	for {
		if resp, err := streamClient.Recv(); err != nil {
			if err != io.EOF {
				fmt.Println(err)
				return
			}
			break
		} else {
			if _, err = dataBuffer.Write(resp.Content); err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	if dataHash, err = util.Sha256Hash2String(dataBuffer.Bytes()); err != nil {
		fmt.Println(err)
		return
	}
	if dataHash != fileHash {
		fmt.Println("file is broken")
		return
	}

	// here, you should put fileContent to a file, example:
	// _ = ioutil.WriteFile(fileName, dataBuffer.Bytes(), 0666)
}

func ExampleServer_FileDelete() {
	c := NewFileDeleteClient(getConn())
	resp, err := c.FileDelete(context.Background(), &FileDeleteRequest{
		Token:            "bf0776c565412060eb93f8f307fae299",
		FileUid:          "556e3b9c936202c9dc67b7ad45530790",
		ForceDeleteIfDir: false,
	})
	fmt.Println(resp, err)
}

func ExampleServer_FileUpdate() {
	c := NewFileUpdateClient(getConn())
	resp, err := c.FileUpdate(context.Background(), &FileUpdateRequest{
		Token:   "bf0776c565412060eb93f8f307fae299",
		FileUid: "556e3b9c936202c9dc67b7ad45530790",
		Path:    "/new/path/to/shield_agents.mp4",
		Hidden:  &wrappers.BoolValue{Value: true},
	})
	fmt.Println(resp, err)
}

func ExampleServer_DirectoryList() {
	c := NewDirectoryListClient(getConn())
	resp, err := c.DirectoryList(context.Background(), &DirectoryListRequest{
		Token: "bf0776c565412060eb93f8f307fae299",
		Sort:  DirectoryListRequest_AscType,
	})
	fmt.Println(resp, err)
}
