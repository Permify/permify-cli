// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"fmt"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	var opts []grpc.DialOption

	if config.CliConfig.CertPath != "" && config.CliConfig.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(config.CliConfig.CertPath, config.CliConfig.CertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load x509 key pair: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
		})))
	} else if config.CliConfig.SslEnabled {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}
