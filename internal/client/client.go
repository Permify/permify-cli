package client

import (
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(endpoint string) (*permify.Client, error) {
	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return client, err
}
