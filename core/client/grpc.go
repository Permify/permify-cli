// Package client handles the permify client to connect with the server
package client

import (
import (
	"context"
	"fmt"
	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	options := []grpc.DialOption{}

	if config.CliConfig.CertPath != "" && config.CliConfig.KeyPath != "" {
		creds, err := credentials.NewClientTLSFromFile(config.CliConfig.CertPath, config.CliConfig.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("could not load tls config: %s", err)
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if config.CliConfig.Token != "" {
		options = append(options, grpc.WithPerRPCCredentials(TokenAuth{
			Token: config.CliConfig.Token,
		}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		options...,
	)
	return client, err
}

type TokenAuth struct {
	Token string
}

func (t TokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.Token,
	}, nil
}

func (t TokenAuth) RequireTransportSecurity() bool {
	return false
}
