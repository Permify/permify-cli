// Package client handles the permify client to connect with the server
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// New initializes a new permify client
func New(endpoint string, token string, certPath string, certKeyPath string) (*permify.Client, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(authUnaryInterceptor(token)),
		grpc.WithStreamInterceptor(authStreamInterceptor(token)),
	}

	transportCredentials, err := transportCredentials(endpoint, certPath, certKeyPath)
	if err != nil {
		return nil, err
	}
	dialOptions = append(dialOptions, grpc.WithTransportCredentials(transportCredentials))

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		dialOptions...,
	)
	return client, err
}

func transportCredentials(endpoint string, certPath string, certKeyPath string) (credentials.TransportCredentials, error) {
	if certPath != "" || certKeyPath != "" {
		if certPath == "" || certKeyPath == "" {
			return nil, fmt.Errorf("both cert_path and cert_key_path must be set")
		}
		cert, err := tls.LoadX509KeyPair(certPath, certKeyPath)
		if err != nil {
			return nil, err
		}
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		return credentials.NewTLS(&tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cert},
		}), nil
	}

	// Keep backwards-compatible default. Many local/dev Permify instances use plaintext gRPC.
	// If the endpoint includes an https scheme, prefer TLS (grpc itself typically uses host:port).
	if strings.HasPrefix(strings.ToLower(endpoint), "https://") {
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		return credentials.NewTLS(&tls.Config{RootCAs: pool}), nil
	}

	return insecure.NewCredentials(), nil
}

func authUnaryInterceptor(token string) grpc.UnaryClientInterceptor {
	value := bearerValue(token)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if value != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", value)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func authStreamInterceptor(token string) grpc.StreamClientInterceptor {
	value := bearerValue(token)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if value != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", value)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func bearerValue(token string) string {
	t := strings.TrimSpace(token)
	if t == "" {
		return ""
	}
	// If user already included "Bearer", don't double-prefix.
	if strings.HasPrefix(strings.ToLower(t), "bearer ") {
		return t
	}
	return "Bearer " + t
}
