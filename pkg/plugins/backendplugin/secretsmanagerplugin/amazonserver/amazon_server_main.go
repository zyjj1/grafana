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
	"github.com/google/uuid"
	pb "github.com/grafana/grafana/pkg/plugins/backendplugin/secretsmanagerplugin"
	grpc "google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
	sm   *secretsmanager.SecretsManager
)

// TODO -- still has risk if the user's key name has this character, maybe encode keyname components as well
const KeySeparator = "///"

// server is used to implement secretsmanagerplugin.RemoteSecretsManagerServer.
type server struct {
	pb.UnimplementedRemoteSecretsManagerServer
}

// Implement server Get func
func (s *server) Get(ctx context.Context, sr *pb.SecretsGetRequest) (*pb.SecretsGetResponse, error) {
	fmt.Println("received secrets GET request", sr)
	out, err := s.PerformGetSecret(ctx, getFormattedSecretName(sr.KeyDescriptor))
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

// Implement server Set func
func (s *server) Set(ctx context.Context, sr *pb.SecretsSetRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets SET request", sr)

	exists, err := s.DoesSecretExist(ctx, sr.KeyDescriptor)
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	formattedSecretName := getFormattedSecretName(sr.KeyDescriptor)
	if exists {
		fmt.Printf("Secret with name %s exists already, updating it", *formattedSecretName)
		_, err = s.PerformUpdateSecret(ctx, formattedSecretName, encryptValue(sr.Value))
	} else {
		fmt.Printf("Secret with name %s does not exist yet, creating a new one", *formattedSecretName)
		_, err = s.PerformCreateSecret(ctx, formattedSecretName, encryptValue(sr.Value))
	}
	if err != nil {
		fmt.Println("Error in set", err.Error())
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

// Implement server Del func
// We want this to be recoverable, but this stops us from being able to create new keys with the same name
// The workaround is to rename this key before soft-deleting it, and then permanently delete the original
func (s *server) Del(ctx context.Context, sr *pb.SecretsDelRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets DEL request", sr)
	// First perform the rename
	softDeleteKey := createSoftDeletionKey(sr.KeyDescriptor)
	_, err := s.Rename(ctx, &pb.SecretsRenameRequest{
		KeyDescriptor: sr.KeyDescriptor,
		NewNamespace:  softDeleteKey.Namespace,
	})
	if err != nil {
		if _, ok := err.(*secretsmanager.ResourceNotFoundException); !ok {
			fmt.Println("Error while renaming key to soft-delete name")
			return &pb.SecretsErrorResponse{
				Error: err.Error(),
			}, err
		}
	}

	// Now soft-delete the new key
	out, err := s.PerformDeleteSecret(ctx, getFormattedSecretName(softDeleteKey), false)
	if err != nil {
		if _, ok := err.(*secretsmanager.ResourceNotFoundException); !ok {
			return &pb.SecretsErrorResponse{
				Error: err.Error(),
			}, err
		}
	} else {
		fmt.Println("Deleted secret with Name", *out.Name)
	}
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

// Implement server Keys func
func (s *server) Keys(ctx context.Context, sr *pb.SecretsKeysRequest) (*pb.SecretsKeysResponse, error) {
	fmt.Println("received secrets KEYS request", sr)
	filter := &secretsmanager.Filter{
		Key:    aws.String("name"),
		Values: []*string{getFormattedKeyPrefix(sr.KeyDescriptor, sr.AllOrganizations)},
	}
	keys, err := s.PerformListKeys(ctx, filter)
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

// Implement server Rename func
func (s *server) Rename(ctx context.Context, sr *pb.SecretsRenameRequest) (*pb.SecretsErrorResponse, error) {
	fmt.Println("received secrets RENAME request", sr)
	// First get the old secret
	getOut, err := s.PerformGetSecret(ctx, getFormattedSecretName(sr.KeyDescriptor))
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}

	// Then create a new secret with the updated name
	createOut, err := s.PerformCreateSecret(ctx, getFormattedSecretUpdatedName(sr.KeyDescriptor, sr.NewNamespace), *getOut.SecretString)
	if err != nil {
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	fmt.Println("Secret created with ARN", createOut.ARN)

	// Then delete the old secret permanently
	_, err = s.PerformDeleteSecret(ctx, getFormattedSecretName(sr.KeyDescriptor), true)
	if err != nil {
		fmt.Println("Error in rename function, failed to delete key with name", getOut.Name)
		return &pb.SecretsErrorResponse{
			Error: err.Error(),
		}, err
	}
	return &pb.SecretsErrorResponse{
		Error: "",
	}, nil
}

// Not part of the interface impl, just helper functions

// Perform GetSecretValueWithContext request to AWS and return the raw response
func (s *server) PerformGetSecret(ctx context.Context, formattedName *string) (*secretsmanager.GetSecretValueOutput, error) {
	return sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: formattedName,
	})
}

// Perform CreateSecretWithContext request to AWS and return the raw response
func (s *server) PerformCreateSecret(ctx context.Context, formattedName *string, encryptedValue string) (*secretsmanager.CreateSecretOutput, error) {
	return sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:         formattedName,
		SecretString: aws.String(encryptedValue),
		Description:  aws.String("Random secret used for testing aws plugin"),
	})
}

// Perform UpdateSecretWithContext request to AWS and return the raw response
func (s *server) PerformUpdateSecret(ctx context.Context, formattedName *string, encryptedValue string) (*secretsmanager.UpdateSecretOutput, error) {
	return sm.UpdateSecretWithContext(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     formattedName,
		SecretString: aws.String(encryptedValue),
	})
}

// Perform GetSecretValueWithContext request to AWS and return the raw response
func (s *server) PerformDeleteSecret(ctx context.Context, formattedName *string, isPermanent bool) (*secretsmanager.DeleteSecretOutput, error) {
	input := &secretsmanager.DeleteSecretInput{
		SecretId: formattedName,
	}
	if isPermanent {
		input.ForceDeleteWithoutRecovery = aws.Bool(true)
	} else {
		input.RecoveryWindowInDays = aws.Int64(7) // This is kind of arbitrary, could just exclude it altogether
	}
	return sm.DeleteSecretWithContext(ctx, input)
}

func (s *server) PerformListKeys(ctx context.Context, nameFilter *secretsmanager.Filter) ([]*pb.Key, error) {
	var keys []*pb.Key = make([]*pb.Key, 0)
	input := &secretsmanager.ListSecretsInput{
		Filters: []*secretsmanager.Filter{nameFilter},
	}
	fmt.Println(input)
	return keys, sm.ListSecretsPagesWithContext(ctx, input, func(out *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		mapSecretEntriesToKeys(out.SecretList, &keys)
		return !lastPage
	})
}

// Perform DescribeSecretWithContext request to AWS. Returns false if there is a ResourceNotFoundException
func (s *server) DoesSecretExist(ctx context.Context, kd *pb.Key) (bool, error) {
	_, err := sm.DescribeSecretWithContext(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: getFormattedSecretName(kd),
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

// Utility functions

// Returns key in the format <ns>///<type>///<org>
func getFormattedSecretName(kd *pb.Key) *string {
	str := fmt.Sprintf("%s%s%s%s%d", sanitizeComponent(kd.Namespace), KeySeparator, sanitizeComponent(kd.Type), KeySeparator, kd.OrgId)
	return &str
}

// Returns key in the format <new-ns>///<type>///<org>
func getFormattedSecretUpdatedName(kd *pb.Key, newNs string) *string {
	str := fmt.Sprintf("%s%s%s%s%d", sanitizeComponent(newNs), KeySeparator, sanitizeComponent(kd.Type), KeySeparator, kd.OrgId)
	return &str
}

// Replaces all instances of the key separator '///' with '-'
func sanitizeComponent(c string) string {
	return strings.ReplaceAll(c, KeySeparator, "-")
}

// Returns search key for ListSecrets call. either <ns>///<type> or <new-ns>///<type>///<org>
func getFormattedKeyPrefix(kd *pb.Key, allOrgs bool) *string {
	str := fmt.Sprintf("%s%s%s%s", sanitizeComponent(kd.Namespace), KeySeparator, sanitizeComponent(kd.Type), KeySeparator)
	if !allOrgs {
		str = fmt.Sprintf("%s%d", str, kd.OrgId)
	}
	return &str
}

func createSoftDeletionKey(kd *pb.Key) *pb.Key {
	return &pb.Key{
		OrgId:     kd.OrgId,
		Type:      kd.Type,
		Namespace: fmt.Sprintf("scheduled-for-deletion-%s/%s", uuid.New().String(), kd.Namespace),
	}
}

// Returns a Key struct containing the namespace, type, and orgId extracted from the provided key
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

// Base64 encodes a value
func encryptValue(val string) string {
	return base64.StdEncoding.EncodeToString([]byte(val))
}

// Base64 decodes a value
func decryptValue(val string) (string, error) {
	rawBytes, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return "", err
	}
	return string(rawBytes), nil
}

// Converts the ListSercrets output into Keys and adds them to the provided slice
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
