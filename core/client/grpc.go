// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client.
func New(endpoint string) (*permify.Client, error) {
	connectionEndpoint := endpoint
	if config.CliCredentials.Endpoint != "" {
		connectionEndpoint = config.CliCredentials.Endpoint
	}
	secureTransport := config.CliConfig.SslEnabled ||
		strings.HasPrefix(config.CliCredentials.Endpoint, "https://") ||
		strings.HasPrefix(connectionEndpoint, "https://") ||
		strings.TrimSpace(config.CliCredentials.CertPath) != "" ||
		strings.TrimSpace(config.CliCredentials.CertKey) != ""
	connectionEndpoint = sanitizeEndpoint(connectionEndpoint)
	if connectionEndpoint == "" {
		return nil, errors.New("permify endpoint is empty")
	}

	transportCredentials, err := transportCredentials(secureTransport)
	if err != nil {
		return nil, err
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),
	}

	token := strings.TrimSpace(config.CliCredentials.Token)
	if token != "" {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(tokenCredentials(token, secureTransport)))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: connectionEndpoint,
		},
		dialOptions...,
	)
	return client, err
}

func sanitizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}

	if !strings.Contains(endpoint, "://") {
		return endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" {
		return endpoint
	}
	if parsed.Host != "" {
		return parsed.Host
	}
	if parsed.Path != "" {
		return strings.TrimLeft(parsed.Path, "/")
	}
	return strings.TrimPrefix(endpoint, parsed.Scheme+"://")
}

func transportCredentials(secureTransport bool) (credentials.TransportCredentials, error) {
	if secureTransport {
		certPath := strings.TrimSpace(config.CliCredentials.CertPath)
		certKey := strings.TrimSpace(config.CliCredentials.CertKey)
		if certPath == "" && certKey == "" {
			return credentials.NewTLS(&tls.Config{}), nil
		}
		if certPath == "" || certKey == "" {
			return nil, fmt.Errorf("both cert path and cert key must be provided")
		}
		if _, err := os.Stat(certPath); err != nil {
			return nil, fmt.Errorf("failed to read cert path: %w", err)
		}
		if _, err := os.Stat(certKey); err != nil {
			return nil, fmt.Errorf("failed to read cert key: %w", err)
		}
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, err
		}
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
		}), nil
	}

	return insecure.NewCredentials(), nil
}

func tokenCredentials(token string, secureTransport bool) credentials.PerRPCCredentials {
	metadata := map[string]string{
		"authorization": "Bearer " + strings.TrimSpace(token),
	}
	if secureTransport {
		return secureTokenCredentials(metadata)
	}
	return nonSecureTokenCredentials(metadata)
}
