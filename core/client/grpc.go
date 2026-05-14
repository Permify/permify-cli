// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	cfg := config.CoreConfig{PermifyURL: endpoint}
	return NewFromConfig(cfg)
}

// NewFromConfig initializes a new permify client using stored CLI configuration.
// It supports optional token auth and optional TLS (client cert + key pair).
func NewFromConfig(cfg config.CoreConfig) (*permify.Client, error) {
	endpoint := normalizeEndpoint(cfg.PermifyURL)
	if endpoint == "" {
		return nil, fmt.Errorf("permify url is empty")
	}

	if (cfg.CertPath == "") != (cfg.CertKeyPath == "") {
		return nil, fmt.Errorf("both cert_path and cert_key_path must be set")
	}

	dialOpts := []grpc.DialOption{}

	transportCreds, err := transportCredentials(cfg)
	if err != nil {
		return nil, err
	}
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(transportCreds))

	if strings.TrimSpace(cfg.Token) != "" {
		token := strings.TrimSpace(cfg.Token)
		headerVal := fmt.Sprintf("Bearer %s", token)
		if transportCreds.Info().SecurityProtocol == "insecure" {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{"authorization": headerVal}))
		} else {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(secureTokenCredentials{"authorization": headerVal}))
		}
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		dialOpts...,
	)
	return client, err
}

func transportCredentials(cfg config.CoreConfig) (credentials.TransportCredentials, error) {
	// Default (backwards compatible): insecure.
	if cfg.CertPath == "" && cfg.CertKeyPath == "" && !tlsEnabled(cfg) {
		return insecure.NewCredentials(), nil
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if cfg.CertPath != "" && cfg.CertKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKeyPath)
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return credentials.NewTLS(tlsCfg), nil
}

func tlsEnabled(cfg config.CoreConfig) bool {
	return cfg.SslEnabled || strings.HasPrefix(strings.ToLower(strings.TrimSpace(cfg.PermifyURL)), "https://")
}

func normalizeEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// If the user provided a URL with scheme, strip scheme/path/query for grpc.Dial endpoint.
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err == nil && u.Host != "" {
			return u.Host
		}
	}

	return raw
}
