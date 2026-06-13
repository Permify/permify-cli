// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
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

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(config.CliConfig.PermifyURL)
	}
	if endpoint == "" {
		return nil, errors.New("permify endpoint missing: provide an endpoint or configure stored credentials")
	}

	opts, err := dialOptions(endpoint, config.CliConfig)
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

func dialOptions(endpoint string, cfg config.CoreConfig) ([]grpc.DialOption, error) {
	transport, err := transportCredentials(endpoint, cfg)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(transport)}

	token := strings.TrimSpace(cfg.Token)
	if token != "" {
		auth := map[string]string{"authorization": fmt.Sprintf("Bearer %s", token)}
		if isSecureEndpoint(endpoint) || cfg.CertPath != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials(auth)))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(auth)))
		}
	}

	return opts, nil
}

func transportCredentials(endpoint string, cfg config.CoreConfig) (credentials.TransportCredentials, error) {
	certPath := strings.TrimSpace(cfg.CertPath)
	certKey := strings.TrimSpace(cfg.CertKey)

	if (certPath == "") != (certKey == "") {
		return nil, errors.New("both cert_path and cert_key must be configured together")
	}

	if certPath != "" {
		if err := fileIsReadable(certPath); err != nil {
			return nil, fmt.Errorf("cert_path: %w", err)
		}
		if err := fileIsReadable(certKey); err != nil {
			return nil, fmt.Errorf("cert_key: %w", err)
		}
		certificate, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, fmt.Errorf("load tls key pair: %w", err)
		}
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS12,
		}), nil
	}

	if isSecureEndpoint(endpoint) {
		return credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}), nil
	}

	return insecure.NewCredentials(), nil
}

func fileIsReadable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	return nil
}

func isSecureEndpoint(endpoint string) bool {
	endpoint = strings.ToLower(strings.TrimSpace(endpoint))
	return strings.HasPrefix(endpoint, "https://") || strings.HasPrefix(endpoint, "grpcs://")
}
