/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	context "context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/grafana/grafana/pkg/plugins/backendplugin/secretsmanagerplugin"
	grpc "google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement secretsmanagerplugin.RemoteSecretsManagerServer.
type server struct {
	pb.UnimplementedRemoteSecretsManagerServer
}

func (s *server) Get(ctx context.Context, sr *pb.SecretsGetRequest) (*pb.SecretsGetResponse, error) {
	fmt.Println("received secrets GET request", sr)
	return &pb.SecretsGetResponse{
		DecryptedValue: "random bs",
		Exists:         true,
		Error:          "",
	}, nil
}
func (s *server) Set(ctx context.Context, sr *pb.SecretsSetRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets SET request", sr)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}
func (s *server) Del(ctx context.Context, sr *pb.SecretsDelRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets DEL request", sr)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}
func (s *server) Keys(ctx context.Context, sr *pb.SecretsKeysRequest) (*pb.SecretsKeysResponse, error) {
	fmt.Println("received secrets KEYS request", sr)
	return &pb.SecretsKeysResponse{
		Error: "",
		Keys: []*pb.Key{{
			OrgId:     69,
			Namespace: "ns",
			Type:      "type",
		}},
	}, nil
}
func (s *server) Rename(ctx context.Context, sr *pb.SecretsRenameRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets RENAME request", sr)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRemoteSecretsManagerServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
