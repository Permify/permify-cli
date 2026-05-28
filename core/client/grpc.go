// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	cliConfig := config.CliConfig
	if endpoint != "" {
		cliConfig.PermifyURL = endpoint
	}
	if cliConfig.PermifyURL == "" {
		return nil, fmt.Errorf("permify endpoint is empty")
	}

	dialOptions, err := newDialOptions(cliConfig)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: cliConfig.PermifyURL,
		},
		dialOptions...,
	)
	return client, err
}

func newDialOptions(cliConfig config.CoreConfig) ([]grpc.DialOption, error) {
	transportCredentials, secureTransport, err := newTransportCredentials(cliConfig.CertPath, cliConfig.CertKey)
	if err != nil {
		return nil, err
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCredentials),
	}

	token := strings.TrimSpace(cliConfig.Token)
	if token == "" {
		return dialOptions, nil
	}

	requestMetadata := map[string]string{
		"authorization": bearerToken(token),
	}
	if secureTransport {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(secureTokenCredentials(requestMetadata)))
	} else {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(requestMetadata)))
	}
	return dialOptions, nil
}

func newTransportCredentials(certPath, certKey string) (credentials.TransportCredentials, bool, error) {
	certPath = strings.TrimSpace(certPath)
	certKey = strings.TrimSpace(certKey)
	if certPath == "" {
		return insecure.NewCredentials(), false, nil
	}

	if certKey == "" {
		transportCredentials, err := credentials.NewClientTLSFromFile(certPath, "")
		if err != nil {
			return nil, false, fmt.Errorf("load TLS certificate %q: %w", certPath, err)
		}
		return transportCredentials, true, nil
	}

	cert, err := tls.LoadX509KeyPair(certPath, certKey)
	if err != nil {
		return nil, false, fmt.Errorf("load TLS certificate pair %q/%q: %w", certPath, certKey, err)
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}), true, nil
}

func bearerToken(token string) string {
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return token
	}
	return "Bearer " + token
}
