// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return false
}

// New initializes a new permify client
func New() (*permify.Client, error) {

	var opts []grpc.DialOption

	if config.CliConfig.CertPath != "" {
		tlsConf := &tls.Config{}
		if config.CliConfig.CertKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(config.CliConfig.CertPath, config.CliConfig.CertKeyPath)
			if err != nil {
				return nil, err
			}
			tlsConf.Certificates = []tls.Certificate{cert}
		} else {
			certPool, err := x509.SystemCertPool()
			if err != nil {
				certPool = x509.NewCertPool()
			}
			b, err := os.ReadFile(config.CliConfig.CertPath)
			if err != nil {
				return nil, err
			}
			certPool.AppendCertsFromPEM(b)
			tlsConf.RootCAs = certPool
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	} else if config.CliConfig.SslEnabled {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if config.CliConfig.Token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(tokenAuth{
			token: config.CliConfig.Token,
		}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: config.CliConfig.PermifyURL,
		},
		opts...,
	)
	return client, err
}
