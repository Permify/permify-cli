// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// New initializes a new permify client
func New(endpoint string, token string, certPath string, certKey string) (*permify.Client, error) {
	var opts []grpc.DialOption

	if certPath != "" {
		var creds credentials.TransportCredentials
		if certKey != "" {
			certificate, err := tls.LoadX509KeyPair(certPath, certKey)
			if err != nil {
				return nil, err
			}
			creds = credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{certificate},
			})
		} else {
			var err error
			creds, err = credentials.NewClientTLSFromFile(certPath, "")
			if err != nil {
				return nil, err
			}
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if token != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
		opts = append(opts, grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			return streamer(ctx, desc, cc, method, opts...)
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
