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
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	pb "github.com/grafana/grafana/pkg/plugins/backendplugin/secretsmanagerplugin"
	grpc "google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
	sm   *secretsmanager.SecretsManager
)

// server is used to implement secretsmanagerplugin.RemoteSecretsManagerServer.
type server struct {
	pb.UnimplementedRemoteSecretsManagerServer
}

func (s *server) Get(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsGetResponse, error) {
	fmt.Println("received secrets GET request", sr)
	out, err := sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: getFormattedSecretName(sr),
	})
	if err != nil {
		return &pb.SecretsGetResponse{
			Error: err.Error(),
		}, err
	}
	decryptedValue, err := decryptValue(*out.SecretString)
	if err != nil {
		return &pb.SecretsGetResponse{
			Error: err.Error(),
		}, err
	}
	return &pb.SecretsGetResponse{
		Exists:         true,
		DecryptedValue: decryptedValue,
	}, nil
}
func (s *server) Set(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets SET request", sr)
	out, err := sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:                        getFormattedSecretName(sr),
		SecretString:                aws.String(encryptValue(sr.Value)),
		Description:                 aws.String("Random secret used for testing aws plugin"),
		ForceOverwriteReplicaSecret: aws.Bool(true),
	})
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	fmt.Println("Secret created with ARN", out.ARN)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}
func (s *server) Del(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets DEL request", sr)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}
func (s *server) Keys(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsKeysResponse, error) {
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
func (s *server) Rename(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets RENAME request", sr)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

func getFormattedSecretName(sr *pb.SecretsRequest) *string {
	str := fmt.Sprintf("%d/%s/%s", sr.OrgId, sr.Namespace, sr.Type)
	return &str
}

func encryptValue(val string) string {
	return base64.StdEncoding.EncodeToString([]byte(val))
}

func decryptValue(val string) (string, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return "", err
	}
	return string(rawBytes), nil
}

func main() {
	flag.Parse()

	mySession := session.Must(session.NewSession())
	sm = secretsmanager.New(mySession, aws.NewConfig().WithRegion("us-east-2").WithLogLevel(aws.LogDebug).WithCredentials(
		credentials.NewSharedCredentials("/Users/mmandrus/dev/aws-cli_accessKeys.csv", "default")))
	// cred file should look like:
	// [default]
	// aws_access_key_id=YOURACCESSKEYID
	// aws_secret_access_key=your/secret/accesskey

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
