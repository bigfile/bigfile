//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package rpc

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/bigfile/bigfile/config"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/internal/util"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func init() {
	isTesting = true
	config.DefaultConfig.Log.File.Enable = false
}
func newPeer() *peer.Peer {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "192.168.0.1:9998")
	return &peer.Peer{
		Addr:     tcpAddr,
		AuthInfo: nil,
	}
}

func newContext(ctx context.Context) context.Context {
	return peer.NewContext(ctx, newPeer())
}

func TestServer_getClientIP(t *testing.T) {
	s := &Server{}
	_, err := s.getClientIP(context.TODO())
	assert.Equal(t, err, ErrGetIPFailed)

	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	ipAddress, err := s.getClientIP(peer.NewContext(context.Background(), &peer.Peer{Addr: tcpAddr}))
	assert.Nil(t, err)
	assert.Equal(t, ipAddress, "127.0.0.1")

	tcpAddr, err = net.ResolveTCPAddr("tcp", "[2000:0:0:0:0:0:0:1]:8080")
	assert.Nil(t, err)
	ipAddress, err = s.getClientIP(peer.NewContext(context.Background(), &peer.Peer{Addr: tcpAddr}))
	assert.Nil(t, err)
	assert.Equal(t, "2000::1", ipAddress)
}

func TestServer_fetchApp(t *testing.T) {
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	_, err = fetchAPP(app.UID, "", trx)
	assert.Equal(t, err, ErrAppSecret)
}

func TestServer_TokenCreate(t *testing.T) {
	s := &Server{}
	app, trx, down, err := models.NewAppForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDbConn = trx
	resp, err := s.TokenCreate(newContext(context.Background()), &TokenCreateRequest{
		AppUid:         app.UID,
		AppSecret:      app.Secret,
		Path:           &wrappers.StringValue{Value: "/random"},
		Ip:             &wrappers.StringValue{Value: "192.168.0.1"},
		Secret:         &wrappers.StringValue{Value: models.RandomWithMd5(22)},
		AvailableTimes: &wrappers.UInt32Value{Value: 1000},
		ReadOnly:       &wrappers.BoolValue{Value: true},
		ExpiredAt:      &timestamp.Timestamp{Seconds: time.Now().Add(10 * time.Minute).Unix()},
	})
	assert.Nil(t, err)
	assert.True(t, resp.RequestId > 0)
	assert.Equal(t, "/random", resp.Token.Path)
	assert.Equal(t, 32, len(resp.Token.Token))
	assert.Equal(t, 32, len(resp.Token.Secret.GetValue()))
	assert.Equal(t, int32(1000), resp.Token.AvailableTimes)
	assert.Equal(t, "192.168.0.1", resp.Token.Ip.GetValue())
	assert.True(t, resp.Token.ReadOnly)
	assert.NotNil(t, resp.Token.ExpiredAt)
	assert.Nil(t, resp.Token.DeletedAt)

	_, err = s.TokenCreate(context.TODO(), &TokenCreateRequest{})
	assert.NotNil(t, err)
	statusError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, statusError.Message(), ErrGetIPFailed.Error())

	_, err = s.TokenCreate(newContext(context.Background()), &TokenCreateRequest{AppUid: app.UID})
	assert.NotNil(t, err)
	statusError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, statusError.Message(), ErrAppSecret.Error())

	_, err = s.TokenCreate(newContext(context.Background()), &TokenCreateRequest{
		AppUid:    app.UID,
		AppSecret: app.Secret,
		Path:      &wrappers.StringValue{Value: "/@@##$$"},
	})
	assert.NotNil(t, err)
	statusError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Contains(t, statusError.Message(), "path is not a legal unix path")
}

func TestServer_TokenUpdate(t *testing.T) {
	s := &Server{}
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDbConn = trx
	resp, err := s.TokenUpdate(newContext(context.Background()), &TokenUpdateRequest{
		AppUid:         token.App.UID,
		AppSecret:      token.App.Secret,
		Token:          token.UID,
		Path:           &wrappers.StringValue{Value: "/new/path"},
		Ip:             &wrappers.StringValue{Value: "192.168.0.2"},
		Secret:         &wrappers.StringValue{Value: models.RandomWithMd5(233)},
		AvailableTimes: &wrappers.UInt32Value{Value: 223},
		ReadOnly:       &wrappers.BoolValue{Value: true},
		ExpiredAt:      &timestamp.Timestamp{Seconds: time.Now().Add(1000 * time.Minute).Unix()},
	})
	assert.Nil(t, err)
	assert.True(t, resp.RequestId > 0)
	assert.Equal(t, "/new/path", resp.Token.Path)
	assert.Equal(t, token.UID, resp.Token.Token)
	assert.Equal(t, 32, len(resp.Token.Secret.GetValue()))
	assert.Equal(t, int32(223), resp.Token.AvailableTimes)
	assert.Equal(t, "192.168.0.2", resp.Token.Ip.GetValue())
	assert.True(t, resp.Token.ReadOnly)
	assert.NotNil(t, resp.Token.ExpiredAt)
	assert.Nil(t, resp.Token.DeletedAt)
}

func TestServer_TokenDelete(t *testing.T) {
	s := &Server{}
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer down(t)
	testDbConn = trx
	resp, err := s.TokenDelete(newContext(context.Background()), &TokenDeleteRequest{
		AppUid:    token.App.UID,
		AppSecret: token.App.Secret,
		Token:     token.UID,
	})
	assert.Nil(t, err)
	assert.True(t, resp.RequestId > 0)
	assert.Equal(t, token.UID, resp.Token.Token)
	assert.NotNil(t, resp.Token.DeletedAt)
}

func TestServer_FileCreate(t *testing.T) {
	const bufSize = 1024 * 1024
	var (
		s            = grpc.NewServer()
		se           *status.Status
		ok           bool
		ctx          = newContext(context.Background())
		err          error
		trx          *gorm.DB
		lis          = bufconn.Listen(bufSize)
		req          = &FileCreateRequest{}
		conn         *grpc.ClientConn
		resp         *FileCreateResponse
		down         func(*testing.T)
		token        *models.Token
		client       FileCreateClient
		rootPath     = models.NewTempDirForTest()
		streamClient FileCreate_FileCreateClient
	)
	token, trx, down, err = models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(rootPath) {
			os.RemoveAll(rootPath)
		}
	}()
	testDbConn = trx
	testRootPath = &rootPath
	RegisterFileCreateServer(s, &Server{})
	go func() { _ = s.Serve(lis) }()
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	assert.Nil(t, err)
	client = NewFileCreateClient(conn)

	// a fake token
	streamClient, err = client.FileCreate(ctx)
	assert.Nil(t, err)
	assert.Nil(t, streamClient.Send(req))
	_, err = streamClient.Recv()
	assert.NotNil(t, err)
	se, ok = status.FromError(err)
	assert.Equal(t, se.Code(), codes.InvalidArgument)
	assert.True(t, ok)
	assert.Equal(t, "record not found", se.Message())

	// token secret error
	secret := models.RandomWithMd5(22)
	assert.Nil(t, trx.Model(token).Update("secret", secret).Error)
	req.Token = token.UID
	streamClient, _ = client.FileCreate(ctx)
	assert.Nil(t, streamClient.Send(req))
	assert.NotNil(t, err)
	_, err = streamClient.Recv()
	se, _ = status.FromError(err)
	assert.Equal(t, ErrTokenSecretWrong.Error(), se.Message())
	req.Secret = &wrappers.StringValue{Value: secret}

	// create dir with content
	req.Path = "/test/create/a/directory"
	req.Operation = &FileCreateRequest_CreateDir{CreateDir: true}
	req.Content = &wrappers.BytesValue{Value: []byte("hello")}
	streamClient, _ = client.FileCreate(ctx)
	assert.Nil(t, streamClient.Send(req))
	_, err = streamClient.Recv()
	assert.NotNil(t, err)
	se, _ = status.FromError(err)
	assert.Equal(t, ErrDirShouldNotHasContent.Error(), se.Message())

	// upload a file
	req.Operation = &FileCreateRequest_None{None: true}
	streamClient, _ = client.FileCreate(ctx)
	assert.Nil(t, streamClient.Send(req))
	resp, err = streamClient.Recv()
	assert.Nil(t, err)
	assert.True(t, resp.RequestId > 0)
	assert.Equal(t, 32, len(resp.File.Uid))
}

func TestServer_FileUpdate(t *testing.T) {
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	testDbConn = trx
	tempDir := models.NewTempDirForTest()
	testRootPath = &tempDir
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	randomBytes := models.Random(222)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	file, err := models.CreateFileFromReader(&token.App, "/random/r.bytes", bytes.NewReader(randomBytes), int8(0), testRootPath, trx)
	assert.Nil(t, err)

	s := Server{}
	resp, err := s.FileUpdate(newContext(context.Background()), &FileUpdateRequest{
		Token:   token.UID,
		FileUid: file.UID,
		Path:    "/new/random.bytes",
		Hidden:  &wrappers.BoolValue{Value: true},
	})
	assert.Nil(t, err)
	assert.Equal(t, "/new/random.bytes", resp.File.Path)
	assert.Equal(t, randomBytesHash, resp.File.Hash.GetValue())
}

func TestServer_FileDelete(t *testing.T) {
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	assert.Nil(t, err)
	testDbConn = trx
	tempDir := models.NewTempDirForTest()
	testRootPath = &tempDir
	defer func() {
		down(t)
		if util.IsDir(tempDir) {
			os.RemoveAll(tempDir)
		}
	}()
	randomBytes := models.Random(222)
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	file, err := models.CreateFileFromReader(&token.App, "/random/r.bytes", bytes.NewReader(randomBytes), int8(0), testRootPath, trx)
	assert.Nil(t, err)

	s := Server{}
	resp, err := s.FileDelete(newContext(context.Background()), &FileDeleteRequest{
		Token:   token.UID,
		FileUid: file.UID,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/random/r.bytes", resp.File.Path)
	assert.Equal(t, randomBytesHash, resp.File.Hash.GetValue())
	assert.NotNil(t, resp.File.GetDeletedAt())
}

func TestServer_FileRead(t *testing.T) {
	// create token
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	rootPath := models.NewTempDirForTest()
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(rootPath) {
			os.RemoveAll(rootPath)
		}
	}()
	testDbConn = trx
	testRootPath = &rootPath

	// create file
	randomBytesSize := models.ChunkSize + 222
	randomBytes := models.Random(uint(randomBytesSize))
	randomBytesHash, err := util.Sha256Hash2String(randomBytes)
	assert.Nil(t, err)
	file, err := models.CreateFileFromReader(&token.App, "/random/r.bytes", bytes.NewReader(randomBytes), int8(0), testRootPath, trx)
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, file.Object.Hash)

	// create server
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	RegisterFileReadServer(s, &Server{})
	go func() { _ = s.Serve(lis) }()
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	ctx := newContext(context.Background())

	// create client
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	assert.Nil(t, err)
	client := NewFileReadClient(conn)
	streamClient, err := client.FileRead(ctx, &FileReadRequest{
		Token:   token.UID,
		FileUid: file.UID,
	})
	assert.Nil(t, err)

	header, err := streamClient.Header()
	assert.Nil(t, err)
	assert.Equal(t, randomBytesHash, header.Get("hash")[0])
	headerSize, err := strconv.Atoi(header.Get("size")[0])
	assert.Nil(t, err)
	assert.Equal(t, headerSize, randomBytesSize)
	dataBuffer := new(bytes.Buffer)
	for {
		if resp, err := streamClient.Recv(); err != nil {
			if err != io.EOF {
				t.Fatal(err)
			} else {
				break
			}
		} else {
			_, err = dataBuffer.Write(resp.Content)
			assert.Nil(t, err)
		}
	}

	dataHash, err := util.Sha256Hash2String(dataBuffer.Bytes())
	assert.Nil(t, err)
	assert.Equal(t, dataBuffer.Len(), randomBytesSize)
	assert.Equal(t, randomBytesHash, dataHash)
}

func TestServer_DirectoryList(t *testing.T) {
	// create token
	token, trx, down, err := models.NewArbitrarilyTokenForTest(nil, t)
	rootPath := models.NewTempDirForTest()
	assert.Nil(t, err)
	defer func() {
		down(t)
		if util.IsDir(rootPath) {
			os.RemoveAll(rootPath)
		}
	}()
	testDbConn = trx
	testRootPath = &rootPath

	// create directories
	number := 18
	for i := 0; i < number; i++ {
		_, err = models.CreateOrGetLastDirectory(&token.App, token.PathWithScope("/test/"+strconv.Itoa(i)), trx)
		assert.Nil(t, err)
	}

	s := Server{}
	resp, err := s.DirectoryList(newContext(context.Background()), &DirectoryListRequest{
		Token: token.UID, SubDir: &wrappers.StringValue{Value: "/test"},
	})
	assert.Nil(t, err)
	assert.Equal(t, int32(number), resp.Total)
	assert.Equal(t, int32(2), resp.Pages)
}
