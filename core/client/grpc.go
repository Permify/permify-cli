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

type Params struct {
	Endpoint    string
	Token       string
	CertPath    string
	CertKeyPath string
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
	if len(trimmed) >= 7 && strings.EqualFold(trimmed[:7], "bearer ") {
		return strings.TrimSpace(trimmed[7:])
	}
	return trimmed
}

func serverName(target string) string {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return target
	}
	return host
}

func dialOptions(params Params) ([]grpc.DialOption, error) {
	parsed := parseEndpoint(params.Endpoint)
	useTLS := parsed.UseTLS

	if (params.CertPath == "") != (params.CertKeyPath == "") {
		return nil, fmt.Errorf("both cert_path and cert_key_path must be set together")
	}

	var options []grpc.DialOption
	if params.CertPath != "" && params.CertKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(params.CertPath, params.CertKeyPath)
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

	token := trimBearerPrefix(params.Token)
	if token != "" {
		if !useTLS {
			return nil, fmt.Errorf("bearer token requires TLS; use an https:// or grpcs:// endpoint, or configure mTLS credentials")
		}
		options = append(options, grpc.WithPerRPCCredentials(secureTokenCredentials{
			"authorization": "Bearer " + token,
		}))
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

func NewFromCLIConfig() (*permify.Client, error) {
	return New(Params{
		Endpoint:    config.CliConfig.PermifyURL,
		Token:       config.CliConfig.Token,
		CertPath:    config.CliConfig.CertPath,
		CertKeyPath: config.CliConfig.CertKeyPath,
	})
}
