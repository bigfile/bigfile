//   Copyright 2019 The bigfile Authors. All rights reserved.
//   Use of this source code is governed by a MIT-style
//   license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bigfile/bigfile/databases/models"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

func init() {
	isTesting = true
}

func newContext(ctx context.Context) context.Context {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "192.168.0.1:9998")
	return peer.NewContext(ctx, &peer.Peer{
		Addr:     tcpAddr,
		AuthInfo: nil,
	})
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
