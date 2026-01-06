// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client with optional persisted credentials.
func New(endpoint string) (*permify.Client, error) {
	storedCreds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	finalEndpoint := endpoint
	if finalEndpoint == "" {
		finalEndpoint = storedCreds.Endpoint
	}
	if finalEndpoint == "" {
		return nil, errors.New("permify endpoint missing: provide an endpoint or store one in ~/.permify/credentials")
	}

	useTLS := isSecureEndpoint(finalEndpoint) || storedCreds.CertPath != "" || storedCreds.CertKey != ""

	dialOptions, err := buildDialOptions(storedCreds, useTLS)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: finalEndpoint,
		},
		dialOptions...,
	)
	return client, err
}

func buildDialOptions(creds Credentials, useTLS bool) ([]grpc.DialOption, error) {
	options := []grpc.DialOption{}

	transportCredentials, err := buildTransportCredentials(creds, useTLS)
	if err != nil {
		return nil, err
	}
	options = append(options, grpc.WithTransportCredentials(transportCredentials))

	if creds.Token != "" {
		bearer := fmt.Sprintf("Bearer %s", creds.Token)
		if transportCredentials.Info().SecurityProtocol != "insecure" {
			options = append(options, grpc.WithPerRPCCredentials(secureTokenCredentials{"authorization": bearer}))
		} else {
			options = append(options, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{"authorization": bearer}))
		}
	}

	return options, nil
}

func buildTransportCredentials(creds Credentials, useTLS bool) (credentials.TransportCredentials, error) {
	if creds.CertPath != "" || creds.CertKey != "" {
		if creds.CertPath == "" || creds.CertKey == "" {
			return nil, errors.New("both certificate path and certificate key must be provided")
		}

		certificate, err := tls.LoadX509KeyPair(creds.CertPath, creds.CertKey)
		if err != nil {
			return nil, fmt.Errorf("load tls key pair: %w", err)
		}

		return credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{certificate}}), nil
	}

	if useTLS {
		return credentials.NewTLS(&tls.Config{}), nil
	}

	return insecure.NewCredentials(), nil
}

func isSecureEndpoint(endpoint string) bool {
	value := strings.ToLower(endpoint)
	return strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "grpcs://")
}
