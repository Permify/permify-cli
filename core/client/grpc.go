// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Credentials holds authentication and TLS configuration for the permify client
type Credentials struct {
	Token    string
	CertPath string
	CertKey  string
}

// New initializes a new permify client with the given endpoint and optional credentials
func New(endpoint string, creds ...Credentials) (*permify.Client, error) {
	var opts []grpc.DialOption

	if len(creds) > 0 && creds[0].Token != "" {
		token := creds[0].Token
		opts = append(opts, grpc.WithPerRPCCredentials(newTokenCredentials(token)))
	}

	if len(creds) > 0 && creds[0].CertPath != "" {
		tlsCreds, err := loadTLSCredentials(creds[0].CertPath, creds[0].CertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
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

// loadTLSCredentials loads TLS certificate and key from file paths
func loadTLSCredentials(certPath, certKeyPath string) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// tokenCredentials implements credentials.PerRPCCredentials for token-based auth
type tokenCredentials struct {
	token string
}

func newTokenCredentials(token string) *tokenCredentials {
	return &tokenCredentials{token: token}
}

func (t *tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t *tokenCredentials) RequireTransportSecurity() bool {
	return false
}
