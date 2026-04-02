package client

import (
	"context"
	"crypto/tls"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return false
}

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	var opts []grpc.DialOption

	if config.CliConfig.CertPath != "" && config.CliConfig.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(config.CliConfig.CertPath, config.CliConfig.CertKey)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
		})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if config.CliConfig.Token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(tokenAuth{
			token: config.CliConfig.Token,
		}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}
