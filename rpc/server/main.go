//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package main

import (
	"log"
	"net"

	"github.com/bigfile/bigfile/rpc"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := &rpc.Server{}
	s := grpc.NewServer()
	rpc.RegisterTokenCreateServer(s, server)
	rpc.RegisterTokenUpdateServer(s, server)
	rpc.RegisterTokenDeleteServer(s, server)
	rpc.RegisterFileCreateServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
