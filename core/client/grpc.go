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

// New initializes a new permify client by retrieving stored credentials
func New(endpoint string) (*permify.Client, error) {
	// 1. Retrieve endpoint: prioritize passed parameter, fall back to stored config or env var
	if endpoint == "" {
		endpoint = config.CliConfig.GetEndpoint()
	}
	if endpoint == "" {
		endpoint = os.Getenv("PERMIFY_ENDPOINT")
	}
	if endpoint == "" {
		return nil, fmt.Errorf("permify endpoint is required: configure via `permctl configure` or PERMIFY_ENDPOINT")
	}

	// 2. Retrieve credentials from storage with environment variable fallbacks
	token := config.CliConfig.Token
	if token == "" {
		token = os.Getenv("PERMIFY_TOKEN")
	}

	certPath := config.CliConfig.CertPath
	if certPath == "" {
		certPath = os.Getenv("PERMIFY_CERT_PATH")
	}

	certKey := config.CliConfig.CertKey
	if certKey == "" {
		certKey = os.Getenv("PERMIFY_CERT_KEY")
	}

	insecureMode := config.CliConfig.Insecure
	if os.Getenv("PERMIFY_INSECURE") == "true" || os.Getenv("PERMIFY_INSECURE") == "1" {
		insecureMode = true
	}

	// 3. Normalize endpoint target for gRPC dial
	isHTTPS := strings.HasPrefix(strings.ToLower(endpoint), "https://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimRight(endpoint, "/")

	// 4. Determine transport security posture
	useTLS := (isHTTPS || config.CliConfig.SslEnabled || certPath != "" || certKey != "") && !insecureMode

	var opts []grpc.DialOption

	if useTLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if certPath != "" && certKey != "" {
			// Mutual TLS: load client certificate and private key
			clientCert, err := tls.LoadX509KeyPair(certPath, certKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate and key (%s, %s): %w", certPath, certKey, err)
			}
			tlsConfig.Certificates = []tls.Certificate{clientCert}
		} else if certPath != "" {
			// Custom CA / Server certificate verification
			caCert, err := os.ReadFile(certPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read cert file (%s): %w", certPath, err)
			}
			caCertPool, err := x509.SystemCertPool()
			if err != nil || caCertPool == nil {
				caCertPool = x509.NewCertPool()
			}
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse certificate from %s", certPath)
			}
			tlsConfig.RootCAs = caCertPool
		} else if certKey != "" {
			return nil, fmt.Errorf("cert_key provided without cert_path: both must be supplied for client TLS")
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

		// Attach bearer token over secure transport
		if token != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials{
				"authorization": fmt.Sprintf("Bearer %s", token),
			}))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

		// Attach bearer token over plaintext transport
		if token != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{
				"authorization": fmt.Sprintf("Bearer %s", token),
			}))
		}
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}

// NewWithConfig initializes a permify client using a specific configuration instance
func NewWithConfig(cfg config.CoreConfig) (*permify.Client, error) {
	oldConfig := config.CliConfig
	config.CliConfig = cfg
	defer func() {
		config.CliConfig = oldConfig
	}()
	return New(cfg.GetEndpoint())
}
