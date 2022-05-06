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
	randomNs   = "random///ns"
	randomNs2  = "random-ns-new"
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
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	orgId := time.Now().UnixMilli()

	// Create a new key
	fmt.Println("Step 1: create a new key with value \"my value\"")
	res, err := c.Set(ctx, &pb.SecretsSetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs,
			Type:      randomType,
		},
		Value: "my value",
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("SET response", res)
	}
	time.Sleep(3 * time.Second)

	// Retrieve the key
	fmt.Println("Step 2: retrieve the key, should have value \"my value\"")
	res2, err := c.Get(ctx, &pb.SecretsGetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs,
			Type:      randomType,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("GET response", res2)
	}
	time.Sleep(3 * time.Second)

	// List the keys for all orgs
	fmt.Println("Step 3: list keys for all orgs, should see our key with orgId", orgId)
	res3, err := c.Keys(ctx, &pb.SecretsKeysRequest{
		KeyDescriptor: &pb.Key{
			Namespace: randomNs,
			Type:      randomType,
		},
		AllOrganizations: true,
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("KEYS response", res3)
	}
	time.Sleep(3 * time.Second)

	// Update the key value
	fmt.Println("Step 4: update the key with new value \"my NEW value\"")
	res4, err := c.Set(ctx, &pb.SecretsSetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs,
			Type:      randomType,
		},
		Value: "my NEW value",
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("SET response", res4)
	}
	time.Sleep(3 * time.Second)

	// Get the key, should be updated
	fmt.Println("Step 5: retrieve the key, which should now have value \"my NEW value\"")
	res5, err := c.Get(ctx, &pb.SecretsGetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs,
			Type:      randomType,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("GET response", res5)
	}
	time.Sleep(3 * time.Second)

	// Rename the key
	fmt.Println("Step 6: rename our key with updated namespace", randomNs2)
	res6, err := c.Rename(ctx, &pb.SecretsRenameRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs,
			Type:      randomType,
		},
		NewNamespace: randomNs2,
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("RENAME response", res6)
	}
	time.Sleep(3 * time.Second)

	// Get the key with the new name, should have new val
	fmt.Println("Step 7: retrieve the key with the new namespace, should still be \"my NEW value\"")
	res7, err := c.Get(ctx, &pb.SecretsGetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs2,
			Type:      randomType,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("GET response", res7)
	}
	time.Sleep(3 * time.Second)

	// List the keys for our org
	fmt.Println("Step 8: list the keys for our org and the new namespace, should have one still")
	res8, err := c.Keys(ctx, &pb.SecretsKeysRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs2,
			Type:      randomType,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("KEYS response", res8)
	}
	time.Sleep(3 * time.Second)

	// Delete our key
	fmt.Println("Step 9: delete the key with the new namespace")
	res9, err := c.Del(ctx, &pb.SecretsDelRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     orgId,
			Namespace: randomNs2,
			Type:      randomType,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("DEL response", res9)
	}
	time.Sleep(3 * time.Second)

	// List the keys for all org one more time, should be empty
	fmt.Println("Step 10: list keys for all orgs, there should be none for orgId", orgId)
	res10, err := c.Keys(ctx, &pb.SecretsKeysRequest{
		KeyDescriptor: &pb.Key{
			Namespace: randomNs,
			Type:      randomType,
		},
		AllOrganizations: true,
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("KEYS response", res10)
	}
	time.Sleep(3 * time.Second)

	// attempt to grab some random val, should get exists false
	fmt.Println("Step 11: get a random key, should give response with exists=false")
	res11, err := c.Get(ctx, &pb.SecretsGetRequest{
		KeyDescriptor: &pb.Key{
			OrgId:     1,
			Namespace: randomNs,
			Type:      randomType,
		},
	})
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("GET response exists =", res11.Exists)
	}

}
