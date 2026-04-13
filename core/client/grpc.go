// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"crypto/x509"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"context"
)

// ClientConfig holds optional connection parameters
type ClientConfig struct {
	Token    string
	CertPath string
	CertKey  string
}

// New initializes a new permify client with optional TLS and token auth
func New(endpoint string, cfg ...ClientConfig) (*permify.Client, error) {
	var opts []grpc.DialOption

	if len(cfg) > 0 && (cfg[0].CertPath != "" || cfg[0].CertKey != "") {
		tlsCfg := &tls.Config{}

		if cfg[0].CertPath != "" && cfg[0].CertKey != "" {
			cert, err := tls.LoadX509KeyPair(cfg[0].CertPath, cfg[0].CertKey)
			if err != nil {
				return nil, err
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}

		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			rootCAs = x509.NewCertPool()
		}
		tlsCfg.RootCAs = rootCAs

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if len(cfg) > 0 && cfg[0].Token != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+cfg[0].Token)
			return invoker(ctx, method, req, reply, cc, callOpts...)
		}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}
