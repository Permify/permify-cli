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

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	// Retrieve credentials from global storage
	token := config.CliConfig.Token
	certPath := config.CliConfig.CertPath
	certKey := config.CliConfig.CertKey

	var creds credentials.TransportCredentials
	if certPath != "" && certKey != "" {
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, err
		}
		creds = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
		})
	} else {
		creds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
			Token:    token,
		},
		opts...,
	)
	return client, err
}
