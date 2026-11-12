// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// ResolveDialOptions validates stored credentials (endpoint, token, certPath, certKey)
// and builds the corresponding gRPC DialOptions for secure or non-secure transport.
func ResolveDialOptions(endpoint, token, certPath, certKey string) (string, []grpc.DialOption, error) {
	resolvedEndpoint := strings.TrimSpace(endpoint)
	if resolvedEndpoint == "" {
		resolvedEndpoint = strings.TrimSpace(config.CliConfig.PermifyURL)
	}
	if resolvedEndpoint == "" {
		return "", nil, errors.New("permify endpoint is missing; please run `permctl configure`")
	}

	var dialOpts []grpc.DialOption
	hasCert := strings.TrimSpace(certPath) != ""
	hasKey := strings.TrimSpace(certKey) != ""

	if hasCert || hasKey {
		if !hasCert || !hasKey {
			return "", nil, errors.New("both cert_path and cert_key must be provided when configuring TLS")
		}
		if _, err := os.Stat(certPath); err != nil {
			return "", nil, fmt.Errorf("certificate file not found at %s: %w", certPath, err)
		}
		if _, err := os.Stat(certKey); err != nil {
			return "", nil, fmt.Errorf("certificate key file not found at %s: %w", certKey, err)
		}
		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return "", nil, fmt.Errorf("failed to load TLS certificate key pair: %w", err)
		}
		tlsCreds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		})
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(tlsCreds))

		if strings.TrimSpace(token) != "" {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(secureTokenCredentials{
				"authorization": formatBearerToken(token),
			}))
		}
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if strings.TrimSpace(token) != "" {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{
				"authorization": formatBearerToken(token),
			}))
		}
	}

	return resolvedEndpoint, dialOpts, nil
}

func formatBearerToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		return trimmed
	}
	return "Bearer " + trimmed
}

// New initializes a new permify client using stored credentials from config.CliConfig.
func New(endpoint string) (*permify.Client, error) {
	storedEndpoint, token, certPath, certKey := config.GetCredentials()
	if strings.TrimSpace(endpoint) == "" {
		endpoint = storedEndpoint
	}

	resolvedEndpoint, dialOpts, err := ResolveDialOptions(endpoint, token, certPath, certKey)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: resolvedEndpoint,
		},
		dialOpts...,
	)
	return client, err
}
