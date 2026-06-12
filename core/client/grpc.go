// Package client handles the permify client to connect with the server
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

// tokenAuth attaches a bearer token, retrieved from storage, to outgoing requests.
type tokenAuth struct {
	token string
}

// GetRequestMetadata returns the authorization metadata for each request.
func (t tokenAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + t.token}, nil
}

// RequireTransportSecurity reports whether the token requires transport security.
func (tokenAuth) RequireTransportSecurity() bool {
	return false
}

// transportCredentials builds the gRPC transport credentials from the stored
// certificate path and key. An insecure connection is used when no certificate
// is configured.
func transportCredentials(certPath, certKey string) (credentials.TransportCredentials, error) {
	if certPath == "" || certKey == "" {
		return insecure.NewCredentials(), nil
	}
	cert, err := tls.LoadX509KeyPair(certPath, certKey)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}}), nil
}

// New initializes a new permify client using the credentials stored in config.
func New(endpoint string) (*permify.Client, error) {
	transportCreds, err := transportCredentials(config.CliConfig.CertPath, config.CliConfig.CertKey)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(transportCreds)}
	if config.CliConfig.Token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(tokenAuth{token: config.CliConfig.Token}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}
