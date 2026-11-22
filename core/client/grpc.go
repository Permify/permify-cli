// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"crypto/x509"
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

// NormalizeEndpoint validates an endpoint string, strips HTTP/gRPC scheme prefixes, and reports whether the scheme implies TLS
func NormalizeEndpoint(endpoint string) (target string, schemeTLS bool, err error) {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return "", false, errors.New("permify endpoint is missing; please run `permctl configure`")
	}

	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "https://"):
		schemeTLS = true
		trimmed = trimmed[len("https://"):]
	case strings.HasPrefix(lower, "grpcs://"):
		schemeTLS = true
		trimmed = trimmed[len("grpcs://"):]
	case strings.HasPrefix(lower, "http://"):
		schemeTLS = false
		trimmed = trimmed[len("http://"):]
	case strings.HasPrefix(lower, "grpc://"):
		schemeTLS = false
		trimmed = trimmed[len("grpc://"):]
	}

	target = strings.TrimSpace(strings.TrimRight(trimmed, "/"))
	if target == "" {
		return "", false, errors.New("permify endpoint is invalid")
	}

	return target, schemeTLS, nil
}

func formatBearerToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(trimmed), "bearer ") {
		return "Bearer " + strings.TrimSpace(trimmed[len("bearer "):])
	}
	return "Bearer " + trimmed
}

// ResolveDialOptions validates stored or provided credentials and builds the gRPC target and dial options
func ResolveDialOptions(endpoint, token, certPath, certKey string) (string, []grpc.DialOption, error) {
	resolvedEndpoint := strings.TrimSpace(endpoint)
	if resolvedEndpoint == "" {
		resolvedEndpoint = strings.TrimSpace(config.CliConfig.PermifyURL)
	}

	target, schemeTLS, err := NormalizeEndpoint(resolvedEndpoint)
	if err != nil {
		return "", nil, err
	}

	certPath = strings.TrimSpace(certPath)
	certKey = strings.TrimSpace(certKey)
	hasCert := certPath != ""
	hasKey := certKey != ""

	var dialOpts []grpc.DialOption
	var useTLS bool

	if hasCert || hasKey {
		if !hasCert || !hasKey {
			return "", nil, errors.New("both cert_path and cert_key must be provided when configuring TLS")
		}

		certInfo, err := os.Stat(certPath)
		if err != nil {
			return "", nil, fmt.Errorf("certificate file not found at %s: %w", certPath, err)
		}
		if certInfo.IsDir() {
			return "", nil, fmt.Errorf("certificate path %s is a directory", certPath)
		}

		keyInfo, err := os.Stat(certKey)
		if err != nil {
			return "", nil, fmt.Errorf("certificate key file not found at %s: %w", certKey, err)
		}
		if keyInfo.IsDir() {
			return "", nil, fmt.Errorf("certificate key path %s is a directory", certKey)
		}

		cert, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return "", nil, fmt.Errorf("failed to load TLS certificate key pair: %w", err)
		}

		certPool, err := x509.SystemCertPool()
		if err != nil || certPool == nil {
			certPool = x509.NewCertPool()
		}
		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read TLS certificate file %s: %w", certPath, err)
		}
		if ok := certPool.AppendCertsFromPEM(certPEM); !ok {
			return "", nil, fmt.Errorf("failed to append certificate from %s to RootCAs", certPath)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      certPool,
			MinVersion:   tls.VersionTLS12,
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		useTLS = true
	} else if schemeTLS || config.CliConfig.SslEnabled {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		useTLS = true
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		useTLS = false
	}

	if strings.TrimSpace(token) != "" {
		formattedToken := formatBearerToken(token)
		if useTLS {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(secureTokenCredentials{
				"authorization": formattedToken,
			}))
		} else {
			dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials{
				"authorization": formattedToken,
			}))
		}
	}

	return target, dialOpts, nil
}

// New initializes a new permify client
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
