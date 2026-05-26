package client

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func New(endpoint, token, certPath, certKey string) (*permify.Client, error) {
	var opts []grpc.DialOption

	if certPath != "" && certKey != "" {
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		caData, err := os.ReadFile(certPath)
		if err == nil {
			certPool.AppendCertsFromPEM(caData)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{"authorization": "Bearer " + token}))
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}
