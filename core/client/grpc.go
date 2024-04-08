// Package client handles the permify client to connect with the server
package client

import (
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		// Todo: Implement secure call with tls certificate
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return client, err
}
