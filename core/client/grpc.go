// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Params contains the connection material needed to initialize a Permify client.
type Params struct {
	Endpoint string
	Token    string
	CertPath string
	CertKey  string
}

type parsedEndpoint struct {
	Target string
	UseTLS bool
}

func parseEndpoint(endpoint string) parsedEndpoint {
	trimmed := strings.TrimSpace(endpoint)
	lower := strings.ToLower(trimmed)

	switch {
	case strings.HasPrefix(lower, "https://"):
		return parsedEndpoint{Target: strings.TrimSpace(trimmed[len("https://"):]), UseTLS: true}
	case strings.HasPrefix(lower, "grpcs://"):
		return parsedEndpoint{Target: strings.TrimSpace(trimmed[len("grpcs://"):]), UseTLS: true}
	case strings.HasPrefix(lower, "http://"):
		return parsedEndpoint{Target: strings.TrimSpace(trimmed[len("http://"):])}
	case strings.HasPrefix(lower, "grpc://"):
		return parsedEndpoint{Target: strings.TrimSpace(trimmed[len("grpc://"):])}
	default:
		return parsedEndpoint{Target: trimmed}
	}
}

func trimBearerPrefix(token string) string {
	trimmed := strings.TrimSpace(token)
	if len(trimmed) >= len("bearer ") && strings.EqualFold(trimmed[:len("bearer ")], "bearer ") {
		return strings.TrimSpace(trimmed[len("bearer "):])
	}
	return trimmed
}

func serverName(target string) string {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return strings.Trim(target, "[]")
	}
	return strings.Trim(host, "[]")
}

func dialOptions(params Params) ([]grpc.DialOption, error) {
	parsed := parseEndpoint(params.Endpoint)
	params.Token = strings.TrimSpace(params.Token)
	params.CertPath = strings.TrimSpace(params.CertPath)
	params.CertKey = strings.TrimSpace(params.CertKey)
	useTLS := parsed.UseTLS

	if (params.CertPath == "") != (params.CertKey == "") {
		return nil, fmt.Errorf("both cert_path and cert_key must be set together")
	}

	var options []grpc.DialOption
	if params.CertPath != "" && params.CertKey != "" {
		cert, err := tls.LoadX509KeyPair(params.CertPath, params.CertKey)
		if err != nil {
			return nil, fmt.Errorf("load client certificate: %w", err)
		}
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			ServerName:   serverName(parsed.Target),
		})))
		useTLS = true
	} else if useTLS {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverName(parsed.Target),
		})))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if token := trimBearerPrefix(params.Token); token != "" {
		authorization := map[string]string{"authorization": "Bearer " + token}
		if useTLS {
			options = append(options, grpc.WithPerRPCCredentials(secureTokenCredentials(authorization)))
		} else {
			options = append(options, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(authorization)))
		}
	}

	return options, nil
}

// New initializes a new permify client
func New(params Params) (*permify.Client, error) {
	parsed := parseEndpoint(params.Endpoint)
	if parsed.Target == "" {
		return nil, fmt.Errorf("endpoint is empty")
	}

	options, err := dialOptions(params)
	if err != nil {
		return nil, err
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: parsed.Target,
		},
		options...,
	)
	return client, err
}

// NewFromCLIConfig initializes a Permify client from the active profile.
func NewFromCLIConfig() (*permify.Client, error) {
	return New(Params{
		Endpoint: config.CliConfig.PermifyURL,
		Token:    config.CliConfig.Token,
		CertPath: config.CliConfig.CertPath,
		CertKey:  config.CliConfig.CertKey,
	})
}
