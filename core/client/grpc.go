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

	dialOptions := []grpc.DialOption{}
	isTLS := strings.HasPrefix(endpoint, "https://")

	if isTLS {
		tlsConfig := &tls.Config{}

		if config.CliConfig.CertPath != "" || config.CliConfig.CertKey != "" {
			if config.CliConfig.CertPath == "" || config.CliConfig.CertKey == "" {
				return nil, fmt.Errorf("both cert_path and cert_key must be set for mTLS")
			}
			certPair, err := tls.LoadX509KeyPair(config.CliConfig.CertPath, config.CliConfig.CertKey)
			if err != nil {
				return nil, fmt.Errorf("load mTLS cert/key: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{certPair}
		}

		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if config.CliConfig.Token != "" {
		if !isTLS {
			return nil, fmt.Errorf("refusing to send token over insecure transport (set an https:// endpoint)")
		}
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(secureTokenCredentials{
			"authorization": "Bearer " + config.CliConfig.Token,
		}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		dialOptions...,
	)
	return client, err
}
