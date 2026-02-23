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
func New(cfg config.CoreConfig) (*permify.Client, error) {
	var opts []grpc.DialOption

	if cfg.CertPath != "" && cfg.CertKey != "" {
		// Load TLS certificate and key
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

		// Use secure token credentials when TLS is enabled
		if cfg.Token != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials{
				"authorization": "Bearer " + cfg.Token,
			}))
		}
	} else {
		// No TLS - use insecure credentials
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

		// Use non-secure token credentials when TLS is not enabled
		if cfg.Token != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{
				"authorization": "Bearer " + cfg.Token,
			}))
		}
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: cfg.PermifyURL,
		},
		opts...,
	)
	return client, err
}
