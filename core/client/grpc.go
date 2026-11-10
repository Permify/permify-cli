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

// New initializes a new permify client using stored credentials and the provided endpoint.
func New(endpoint string) (*permify.Client, error) {
	cfg := config.CliConfig
	if strings.TrimSpace(endpoint) != "" {
		cfg.PermifyURL = strings.TrimSpace(endpoint)
	}
	if strings.TrimSpace(cfg.PermifyURL) == "" {
		return nil, fmt.Errorf("permify endpoint is empty")
	}
	return NewWithConfig(cfg)
}

// NewWithConfig initializes a new permify client from the provided CoreConfig.
func NewWithConfig(cfg config.CoreConfig) (*permify.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	dialOpts, err := buildDialOptions(cfg)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: cfg.PermifyURL,
		},
		dialOpts...,
	)
	return client, err
}

func buildDialOptions(cfg config.CoreConfig) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	useTLS := cfg.SslEnabled ||
		strings.HasPrefix(strings.ToLower(cfg.PermifyURL), "https://") ||
		strings.HasPrefix(strings.ToLower(cfg.PermifyURL), "grpcs://") ||
		(cfg.CertPath != "" && cfg.CertKey != "")

	if cfg.CertPath != "" && cfg.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
		}
		tlsCfg := &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else if useTLS {
		tlsCfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if strings.TrimSpace(cfg.Token) != "" {
		bearer := "Bearer " + strings.TrimPrefix(strings.TrimSpace(cfg.Token), "Bearer ")
		if useTLS {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials{
				"authorization": bearer,
			}))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{
				"authorization": bearer,
			}))
		}
	}

	return opts, nil
}
