// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client using stored credentials (endpoint, token, cert path, and cert key).
func New(endpoint string) (*permify.Client, error) {
	resolvedEndpoint := strings.TrimSpace(endpoint)
	if resolvedEndpoint == "" {
		resolvedEndpoint = strings.TrimSpace(config.CliConfig.PermifyURL)
	}
	if resolvedEndpoint == "" {
		return nil, errors.New("permify endpoint is missing: run `permctl configure` or provide a valid endpoint")
	}

	dialOpts := []grpc.DialOption{}
	certPath := strings.TrimSpace(config.CliConfig.CertPath)
	certKey := strings.TrimSpace(config.CliConfig.CertKey)
	token := strings.TrimSpace(config.CliConfig.Token)

	useTLS := false
	if certPath != "" || certKey != "" {
		if certPath != "" && certKey != "" {
			keyPair, err := tls.LoadX509KeyPair(certPath, certKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client TLS certificate and key: %w", err)
			}
			tlsCfg := &tls.Config{
				Certificates: []tls.Certificate{keyPair},
				MinVersion:   tls.VersionTLS12,
			}
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
			useTLS = true
		} else if certPath != "" {
			creds, err := credentials.NewClientTLSFromFile(certPath, "")
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS certificate from %s: %w", certPath, err)
			}
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
			useTLS = true
		} else {
			return nil, errors.New("cert_path is required when cert_key is specified")
		}
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if token != "" {
		authHeader := map[string]string{
			"authorization": "Bearer " + token,
		}
		if useTLS {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(secureTokenCredentials(authHeader)))
		} else {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(authHeader)))
		}
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: resolvedEndpoint,
		},
		dialOpts...,
	)
	return client, err
}
