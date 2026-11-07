// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client using stored credentials if configured
func New(endpoint string) (*permify.Client, error) {
	if endpoint == "" && config.CliConfig.PermifyURL != "" {
		endpoint = config.CliConfig.PermifyURL
	}

	var dialOpts []grpc.DialOption
	if config.CliConfig.CertPath != "" && config.CliConfig.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(config.CliConfig.CertPath, config.CliConfig.CertKey)
		if err != nil {
			return nil, err
		}
		tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else if config.CliConfig.CertPath != "" {
		transportCreds, err := credentials.NewClientTLSFromFile(config.CliConfig.CertPath, "")
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(transportCreds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cfg := permify.Config{
		Endpoint: endpoint,
	}
	if config.CliConfig.Token != "" {
		cfg.Token = config.CliConfig.Token
	}

	client, err := permify.NewClient(cfg, dialOpts...)
	return client, err
}
