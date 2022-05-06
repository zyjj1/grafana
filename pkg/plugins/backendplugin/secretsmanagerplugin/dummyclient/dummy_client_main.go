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

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/grafana/grafana/pkg/plugins/backendplugin/secretsmanagerplugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRemoteSecretsManagerClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.Get(ctx, &pb.SecretsGetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     100,
			Namespace: "random ns",
			Type:      "random type",
		},
	})
	fmt.Println("GET response", res)
	res2, err := c.Set(ctx, &pb.SecretsSetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     100,
			Namespace: "random ns",
			Type:      "random type",
		},
		Value: "random value",
	})
	fmt.Println("SET response", res2)
	res3, err := c.Del(ctx, &pb.SecretsDelRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     100,
			Namespace: "random ns",
			Type:      "random type",
		},
	})
	fmt.Println("DEL response", res3)
	res4, err := c.Rename(ctx, &pb.SecretsRenameRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     100,
			Namespace: "random ns",
			Type:      "random type",
		},
		NewNamespace: "random new ns",
	})
	fmt.Println("RENAME response", res4)
	res5, err := c.Keys(ctx, &pb.SecretsKeysRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     100,
			Namespace: "random ns",
			Type:      "random type",
		},
		AllOrganizations: false,
	})
	fmt.Println("KEYS response", res5)
}
