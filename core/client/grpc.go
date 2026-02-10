// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"fmt"
	"strings"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client using stored credentials for
// TLS transport and bearer token authentication when available.
func New(endpoint string) (*permify.Client, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("load credentials: %w", err)
	}

	opts, err := buildDialOptions(creds, endpoint)
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

// buildDialOptions constructs gRPC dial options based on stored credentials.
func buildDialOptions(creds Credentials, endpoint string) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	tlsEnabled, err := appendTransportCredentials(&opts, creds, endpoint)
	if err != nil {
		return nil, err
	}

	appendTokenCredentials(&opts, creds, tlsEnabled)

	return opts, nil
}

// appendTransportCredentials adds TLS or insecure transport to the dial options.
// Returns true if TLS is enabled.
func appendTransportCredentials(opts *[]grpc.DialOption, creds Credentials, endpoint string) (bool, error) {
	if creds.CertPath != "" && creds.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(creds.CertPath, creds.CertKey)
		if err != nil {
			return false, fmt.Errorf("load TLS key pair: %w", err)
		}
		tlsCreds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
		})
		*opts = append(*opts, grpc.WithTransportCredentials(tlsCreds))
		return true, nil
	}

	if isSecureEndpoint(endpoint) {
		tlsCreds := credentials.NewTLS(&tls.Config{})
		*opts = append(*opts, grpc.WithTransportCredentials(tlsCreds))
		return true, nil
	}

	*opts = append(*opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return false, nil
}

// appendTokenCredentials adds bearer token per-RPC credentials to the dial options
// using the appropriate credential type based on whether TLS is enabled.
func appendTokenCredentials(opts *[]grpc.DialOption, creds Credentials, tlsEnabled bool) {
	if creds.Token == "" {
		return
	}

	bearer := fmt.Sprintf("Bearer %s", creds.Token)
	if tlsEnabled {
		*opts = append(*opts, grpc.WithPerRPCCredentials(secureTokenCredentials{
			"authorization": bearer,
		}))
	} else {
		*opts = append(*opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{
			"authorization": bearer,
		}))
	}
}

// isSecureEndpoint checks if the endpoint uses a secure protocol.
func isSecureEndpoint(endpoint string) bool {
	lower := strings.ToLower(endpoint)
	return strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "grpcs://")
}
