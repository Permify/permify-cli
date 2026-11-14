// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type tokenCredentials struct {
	token string
}

func (t tokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t tokenCredentials) RequireTransportSecurity() bool {
	return false
}

// New initializes a new permify client using stored credentials from configuration
func New(endpoint string) (*permify.Client, error) {
	resolvedEndpoint := strings.TrimSpace(endpoint)
	if resolvedEndpoint == "" {
		resolvedEndpoint = strings.TrimSpace(config.CliConfig.PermifyURL)
	}
	if resolvedEndpoint == "" {
		return nil, errors.New("permify endpoint is missing; please configure endpoint using `permctl configure`")
	}

	certPath := strings.TrimSpace(config.CliConfig.CertPath)
	certKey := strings.TrimSpace(config.CliConfig.CertKey)
	token := strings.TrimSpace(config.CliConfig.Token)

	dialOpts, err := buildDialOptions(certPath, certKey, token)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: resolvedEndpoint,
		},
		dialOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create permify client: %w", err)
	}
	return client, nil
}

func buildDialOptions(certPath, certKey, token string) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	if certKey != "" && certPath == "" {
		return nil, errors.New("cert_path is required when cert_key is provided")
	}

	if certPath != "" && certKey != "" {
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate key pair (%s, %s): %w", certPath, certKey, err)
		}
		caBytes, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate file %s: %w", certPath, err)
		}
		certPool, err := x509.SystemCertPool()
		if err != nil || certPool == nil {
			certPool = x509.NewCertPool()
		}
		certPool.AppendCertsFromPEM(caBytes)
		tlsCreds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS12,
		})
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	} else if certPath != "" {
		tlsCreds, err := credentials.NewClientTLSFromFile(certPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate from %s: %w", certPath, err)
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(tokenCredentials{token: token}))
	}

	return opts, nil
}
