// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	permify "github.com/Permify/permify-go/v1"
	"github.com/Permify/permify-cli/core/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	var opts []grpc.DialOption

	// Handle TLS Configuration
	if config.CliConfig.CertPath != "" {
		// Load client certificate if provided
		var creds credentials.TransportCredentials
		if config.CliConfig.CertKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(config.CliConfig.CertPath, config.CliConfig.CertKeyPath)
			if err != nil {
				return nil, err
			}
			creds = credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
			})
		} else {
			// Load CA cert
			caCert, err := os.ReadFile(config.CliConfig.CertPath)
			if err != nil {
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			creds = credentials.NewTLS(&tls.Config{
				RootCAs: caCertPool,
			})
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Handle Token Authentication
	// Check environment variable first for maximum flexibility
	token := os.Getenv("PERMIFY_TOKEN")
	if token == "" {
		token = config.CliConfig.Token
	}

	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(tokenAuth{
			token: token,
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
