// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type bearerTokenCredentials struct {
	token      string
	requireTLS bool
}

func (b bearerTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + b.token,
	}, nil
}

func (b bearerTokenCredentials) RequireTransportSecurity() bool {
	return b.requireTLS
}

// buildTransportCredentials constructs gRPC transport credentials from stored CLI config.
func buildTransportCredentials(creds config.CoreConfig) (credentials.TransportCredentials, error) {
	if creds.CertPath != "" && creds.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(creds.CertPath, creds.CertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}), nil
	}
	if creds.CertPath != "" {
		tlsCreds, err := credentials.NewClientTLSFromFile(creds.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load CA certificate from %s: %w", creds.CertPath, err)
		}
		return tlsCreds, nil
	}
	return insecure.NewCredentials(), nil
}

// New initializes a new permify client using stored endpoint, token, cert path, and cert key credentials.
func New(endpoint string) (*permify.Client, error) {
	creds, err := config.GetStoredCredentials(endpoint)
	if err != nil {
		return nil, err
	}

	transportCreds, err := buildTransportCredentials(creds)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
	}
	if creds.Token != "" {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(bearerTokenCredentials{
			token:      creds.Token,
			requireTLS: creds.SslEnabled,
		}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: creds.PermifyURL,
		},
		dialOpts...,
	)
	return client, err
}
