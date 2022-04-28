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

const (
	randomNs   = "random-ns"
	randomType = "random-type"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.Set(ctx, &pb.SecretsRequest{
		OrgId:     123,
		Namespace: randomNs,
		Type:      randomType,
		Value:     "my value",
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("SET response", res)
	}

	time.Sleep(3 * time.Second)
	res2, err := c.Get(ctx, &pb.SecretsRequest{
		OrgId:     123,
		Namespace: randomNs,
		Type:      randomType,
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("GET response", res2)
	}

	time.Sleep(3 * time.Second)
	res3, err := c.Keys(ctx, &pb.SecretsRequest{
		OrgId:     123,
		Namespace: randomNs,
		Type:      randomType,
	})
	fmt.Println("KEYS response", res3)

	// res3, err := c.Del(ctx, &pb.SecretsRequest{
	// 	OrgId:     123,
	// 	Namespace: randomNs,
	// 	Type:      randomType,
	// })
	// fmt.Println("DEL response", res3)
	// res4, err := c.Rename(ctx, &pb.SecretsRequest{
	// 	OrgId:     123,
	// 	Namespace: randomNs,
	// 	Type:      randomType,
	// })
	// fmt.Println("RENAME response", res4)
	// res5, err := c.Keys(ctx, &pb.SecretsRequest{
	// 	OrgId:     123,
	// 	Namespace: randomNs,
	// 	Type:      randomType,
	// })
	// fmt.Println("KEYS response", res5)
}
