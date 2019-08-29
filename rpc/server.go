//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc/peer"
)

var (
	// ErrGetIPFailed represent that get ip failed
	ErrGetIPFailed = errors.New("[getClientIP] invoke FromContext() failed")

	// ErrEmptyClientIP represent that the input ip is empty
	ErrEmptyClientIP = errors.New("[getClientIP] peer.Addr is nil")
)

// Server is used to create a rpc server
type Server struct{}

func (s *Server) getClientIP(ctx context.Context) (string, error) {
	var (
		pr *peer.Peer
		ok bool
	)
	if pr, ok = peer.FromContext(ctx); !ok {
		return "", ErrGetIPFailed
	}
	if pr.Addr == net.Addr(nil) {
		return "", ErrEmptyClientIP
	}
	return pr.Addr.String(), nil
}

// TokenCreate is used to implement create token service
func (s *Server) TokenCreate(ctx context.Context, req *TokenCreateRequest) (*TokenCreateResponse, error) {
	fmt.Println(s.getClientIP(ctx))
	return &TokenCreateResponse{}, nil
}
