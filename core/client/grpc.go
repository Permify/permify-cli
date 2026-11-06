// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"crypto/x509"
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
	transportCreds, err := transportCredentials(config.CliConfig)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(transportCreds)}
	if tokenCreds := tokenCredentials(config.CliConfig); tokenCreds != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(tokenCreds))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: normalizeEndpoint(endpoint),
		},
		opts...,
	)
	return client, err
}

func transportCredentials(cfg config.CoreConfig) (credentials.TransportCredentials, error) {
	if cfg.CertPath == "" {
		if strings.TrimSpace(cfg.CertKeyPath) != "" {
			return nil, fmt.Errorf("cert key path requires cert path")
		}
		if cfg.SslEnabled {
			return credentials.NewTLS(&tls.Config{
				MinVersion: tls.VersionTLS12,
			}), nil
		}
		return insecure.NewCredentials(), nil
	}

	if cfg.CertKeyPath == "" {
		return credentials.NewClientTLSFromFile(cfg.CertPath, "")
	}

	certPool := x509.NewCertPool()
	certPEM, err := os.ReadFile(cfg.CertPath)
	if err != nil {
		return nil, err
	}
	if !certPool.AppendCertsFromPEM(certPEM) {
		return nil, fmt.Errorf("failed to parse certificate at %s", cfg.CertPath)
	}

	clientCertificate, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKeyPath)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCertificate},
	}), nil
}

func tokenCredentials(cfg config.CoreConfig) credentials.PerRPCCredentials {
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil
	}

	metadata := map[string]string{
		"authorization": bearerToken(token),
	}
	if cfg.SslEnabled || cfg.CertPath != "" {
		return secureTokenCredentials(metadata)
	}
	return nonSecureTokenCredentials(metadata)
}

func bearerToken(token string) string {
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token
	}
	return fmt.Sprintf("Bearer %s", token)
}

func normalizeEndpoint(endpoint string) string {
	withoutHTTPS := strings.TrimPrefix(endpoint, "https://")
	return strings.TrimPrefix(withoutHTTPS, "http://")
}
