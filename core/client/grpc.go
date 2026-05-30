// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// tokenCredentials implements grpc.PerRPCCredentials for bearer token auth.
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

// Options holds optional connection parameters loaded from stored credentials.
type Options struct {
	Token       string
	CertPath    string
	CertKeyPath string
}

// New initializes a new permify client using the provided endpoint and optional
// credentials. When CertPath and CertKeyPath are set, mTLS is used. When only
// Token is set, a bearer token is attached to every request.
func New(endpoint string, opts Options) (*permify.Client, error) {
	var dialOpts []grpc.DialOption

	hasCert := opts.CertPath != "" && opts.CertKeyPath != ""

	if hasCert {
		cert, err := tls.LoadX509KeyPair(opts.CertPath, opts.CertKeyPath)
		if err != nil {
			return nil, fmt.Errorf("loading TLS key pair: %w", err)
		}

		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      x509.NewCertPool(),
		}

		// Attempt to append system root CAs; proceed without them if unavailable.
		if pool, err := x509.SystemCertPool(); err == nil {
			tlsCfg.RootCAs = pool
		}

		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if opts.Token != "" {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(tokenCredentials{token: opts.Token}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		dialOpts...,
	)
	return client, err
}

// NewFromEnv initializes a permify client, preferring environment variables
// PERMIFY_TOKEN, PERMIFY_CERT_PATH, and PERMIFY_CERT_KEY_PATH over the stored
// options when they are set.
func NewFromEnv(endpoint string, stored Options) (*permify.Client, error) {
	opts := stored
	if v := os.Getenv("PERMIFY_TOKEN"); v != "" {
		opts.Token = v
	}
	if v := os.Getenv("PERMIFY_CERT_PATH"); v != "" {
		opts.CertPath = v
	}
	if v := os.Getenv("PERMIFY_CERT_KEY_PATH"); v != "" {
		opts.CertKeyPath = v
	}
	return New(endpoint, opts)
}
