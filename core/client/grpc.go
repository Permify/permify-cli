// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"crypto/x509"
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
	if endpoint == "" {
		endpoint = config.CliConfig.PermifyURL
	}

	var dialOpts []grpc.DialOption

	if config.CliConfig.SslEnabled {
		tlsConfig, err := buildTLSConfig(config.CliConfig.CertPath, config.CliConfig.CertKey)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if token := strings.TrimSpace(config.CliConfig.Token); token != "" {
		md := map[string]string{
			"authorization": formatBearerToken(token),
		}
		if config.CliConfig.SslEnabled {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(secureTokenCredentials(md)))
		} else {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(md)))
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

func formatBearerToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token
	}
	return "Bearer " + token
}

func buildTLSConfig(certPath, certKey string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	certPath = strings.TrimSpace(certPath)
	certKey = strings.TrimSpace(certKey)

	// If both cert+key are provided, use them as a client certificate (mTLS).
	if certPath != "" && certKey != "" {
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		return tlsConfig, nil
	}

	// If only a cert path is provided, treat it as the CA bundle for verifying the server.
	if certPath != "" {
		caPem, err := os.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM(caPem); ok {
			tlsConfig.RootCAs = rootCAs
		}
	}

	return tlsConfig, nil
}
