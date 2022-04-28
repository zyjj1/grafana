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
	"strconv"
	"strings"

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

// TODO -- better to make this a boolean flag in the request
const AllOrganizations = -1

// TODO -- still has risk if the user's key name has this character, maybe encode keyname components as well
const KeySeparator = "///"

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
		switch e := err.(type) {
		case *secretsmanager.ResourceNotFoundException:
			return &pb.SecretsGetResponse{
				Exists:         false,
				DecryptedValue: "",
			}, nil
		default:
			return &pb.SecretsGetResponse{
				Error: e.Error(),
			}, e
		}
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

	exists, err := s.SecretExists(ctx, sr)
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	if exists {
		fmt.Printf("Secret with name %s exists already, updating it", *getFormattedSecretName(sr))
		return s.UpdateSecret(ctx, sr)
	} else {
		fmt.Printf("Secret with name %s does not exist yet, creating a new one", *getFormattedSecretName(sr))
		return s.CreateSecret(ctx, sr)
	}
}
func (s *server) Del(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets DEL request", sr)
	out3, err := sm.DeleteSecretWithContext(ctx, &secretsmanager.DeleteSecretInput{
		// ForceDeleteWithoutRecovery: aws.Bool(true),
		RecoveryWindowInDays: aws.Int64(7), // TODO Determine if we can just force delete without recovery
		SecretId:             getFormattedSecretName(sr),
	})
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	fmt.Println("Deleted secret with Name", out3.Name)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

func (s *server) Keys(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsKeysResponse, error) {
	fmt.Println("received secrets KEYS request", sr)
	var keys []*pb.Key = make([]*pb.Key, 0)
	input := &secretsmanager.ListSecretsInput{
		Filters: []*secretsmanager.Filter{{
			Key:    aws.String("name"),
			Values: []*string{getFormattedKeyPrefix(sr)},
		}},
	}
	err := sm.ListSecretsPagesWithContext(ctx, input, func(out *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		mapSecretEntriesToKeys(out.SecretList, &keys)
		return !lastPage
	})
	if err != nil {
		return &pb.SecretsKeysResponse{
			Error: err.Error(),
		}, err
	}
	return &pb.SecretsKeysResponse{
		Error: "",
		Keys:  keys,
	}, nil
}

// TODO refactor so that we have more function re-use
func (s *server) Rename(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets RENAME request", sr)
	// First get the old secret
	out, err := sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: getFormattedSecretName(sr),
	})
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}

	// Then create a new secret with the updated name
	out2, err := sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:         getFormattedSecretUpdatedName(sr),
		SecretString: out.SecretString,
		Description:  aws.String("Random secret used for testing aws plugin"),
	})
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	fmt.Println("Secret created with ARN", out2.ARN)

	// Then delete the old secret
	delRes, err := s.Del(ctx, sr)
	if err != nil {
		fmt.Println("Error in rename function, failed to delete key with name", out.Name)
	}
	return delRes, err
}

// Not part of the interface impl, just helper functions

func (s *server) CreateSecret(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	out, err := sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:         getFormattedSecretName(sr),
		SecretString: aws.String(encryptValue(sr.Value)),
		Description:  aws.String("Random secret used for testing aws plugin"),
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

func (s *server) UpdateSecret(ctx context.Context, sr *pb.SecretsRequest) (*pb.SecretsErrorResponse, error) {
	out, err := sm.UpdateSecretWithContext(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     getFormattedSecretName(sr),
		SecretString: aws.String(encryptValue(sr.Value)),
	})
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	fmt.Println("Secret updated with ARN", out.ARN)
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

func (s *server) SecretExists(ctx context.Context, sr *pb.SecretsRequest) (bool, error) {
	_, err := sm.DescribeSecretWithContext(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: getFormattedSecretName(sr),
	})

	if err != nil {
		switch e := err.(type) {
		case *secretsmanager.ResourceNotFoundException:
			return false, nil
		default:
			return false, e
		}
	}
	return true, nil
}

func getFormattedSecretName(sr *pb.SecretsRequest) *string {
	str := fmt.Sprintf("%s%s%s%s%d", sr.Namespace, KeySeparator, sr.Type, KeySeparator, sr.OrgId)
	return &str
}

func getFormattedSecretUpdatedName(sr *pb.SecretsRequest) *string {
	str := fmt.Sprintf("%s%s%s%s%d", sr.NewNamespace, KeySeparator, sr.Type, KeySeparator, sr.OrgId)
	return &str
}

func getFormattedKeyPrefix(sr *pb.SecretsRequest) *string {
	str := fmt.Sprintf("%s%s%s%s", sr.Namespace, KeySeparator, sr.Type, KeySeparator)
	if sr.OrgId != AllOrganizations {
		str = fmt.Sprintf("%s%d", str, sr.OrgId)
	}
	return &str
}

func getKeyForFormattedSecretName(name string) (*pb.Key, error) {
	sp := strings.Split(name, KeySeparator)
	ns := sp[0]
	typ := sp[1]
	org, err := strconv.ParseInt(sp[2], 10, 64)
	if err != nil {
		return nil, err
	}
	return &pb.Key{
		Namespace: ns,
		Type:      typ,
		OrgId:     org,
	}, nil
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

func mapSecretEntriesToKeys(sl []*secretsmanager.SecretListEntry, keysPtr *[]*pb.Key) {
	keys := *keysPtr
	for _, entry := range sl {
		k, err := getKeyForFormattedSecretName(*entry.Name)
		if err != nil {
			fmt.Printf("Error converting secret entry to Key: %s", err.Error())
		} else {
			*keysPtr = append(keys, k)
		}
	}

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
