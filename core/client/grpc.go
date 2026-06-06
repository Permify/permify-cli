// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	if endpoint == "" {
		endpoint = config.CliConfig.PermifyURL
	}
	if endpoint == "" {
		return nil, fmt.Errorf("permify endpoint is missing")
	}

	opts, err := dialOptions(config.CliConfig)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}

func dialOptions(cfg config.CoreConfig) ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{}

	transportCredentials, err := transportCredentials(cfg)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithTransportCredentials(transportCredentials))

	if cfg.Token != "" {
		token := cfg.Token
		if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = "Bearer " + token
		}
		if cfg.SslEnabled || cfg.CertPath != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials{"authorization": token}))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{"authorization": token}))
		}
	}

	return opts, nil
}

func transportCredentials(cfg config.CoreConfig) (credentials.TransportCredentials, error) {
	if cfg.CertPath != "" && cfg.CertKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate pair: %w", err)
		}
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}), nil
	}

	if cfg.CertKeyPath != "" {
		return nil, fmt.Errorf("cert_key_path requires cert_path")
	}

	if cfg.CertPath != "" {
		return credentials.NewClientTLSFromFile(cfg.CertPath, "")
	}

	if cfg.SslEnabled {
		return credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}), nil
	}

	return insecure.NewCredentials(), nil
}
