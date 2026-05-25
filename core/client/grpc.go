// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type permifyConfig = permify.Config
type permifyClient = *permify.Client

type clientFactory func(permifyConfig, ...grpc.DialOption) (permifyClient, error)

type bearerTokenCredentials struct {
	token                    string
	requireTransportSecurity bool
}

func (c bearerTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + c.token}, nil
}

func (c bearerTokenCredentials) RequireTransportSecurity() bool {
	return c.requireTransportSecurity
}

// DialOptions builds gRPC dial options from the persisted CLI configuration.
func DialOptions(cfg config.CoreConfig) ([]grpc.DialOption, error) {
	requireTransportSecurity := strings.HasPrefix(cfg.PermifyURL, "https") || cfg.SslEnabled

	opts := []grpc.DialOption{}
	transportCredentials, err := transportCredentials(cfg, requireTransportSecurity)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithTransportCredentials(transportCredentials))

	if cfg.Token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(bearerTokenCredentials{
			token:                    cfg.Token,
			requireTransportSecurity: requireTransportSecurity,
		}))
	}

	return opts, nil
}

func transportCredentials(cfg config.CoreConfig, secure bool) (credentials.TransportCredentials, error) {
	if !secure {
		return insecure.NewCredentials(), nil
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.CertPath != "" || cfg.CertKey != "" {
		if cfg.CertPath == "" || cfg.CertKey == "" {
			return nil, fmt.Errorf("both cert path and cert key must be configured for TLS client certificates")
		}
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return credentials.NewTLS(tlsConfig), nil
}

func newFromConfig(cfg config.CoreConfig, factory clientFactory) (*permify.Client, error) {
	opts, err := DialOptions(cfg)
	if err != nil {
		return nil, err
	}
	return factory(permify.Config{Endpoint: cfg.PermifyURL}, opts...)
}

// New initializes a new permify client with an explicit endpoint.
func New(endpoint string) (*permify.Client, error) {
	return NewFromConfig(config.CoreConfig{PermifyURL: endpoint})
}

// NewFromConfig initializes a new permify client from persisted CLI configuration.
func NewFromConfig(cfg config.CoreConfig) (*permify.Client, error) {
	return newFromConfig(cfg, permify.NewClient)
}
