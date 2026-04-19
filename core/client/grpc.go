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

// parsedGRPCEndpoint is the dial target plus whether the URL explicitly required TLS to the server
// (https:// or grpcs://). Stripping the scheme must not silently downgrade to plaintext when a token is used.
type parsedGRPCEndpoint struct {
	Target          string
	ServerTLSScheme bool
}

func parseGRPCEndpoint(endpoint string) parsedGRPCEndpoint {
	e := strings.TrimSpace(endpoint)
	lower := strings.ToLower(e)
	if strings.HasPrefix(lower, "https://") {
		return parsedGRPCEndpoint{Target: strings.TrimSpace(e[len("https://"):]), ServerTLSScheme: true}
	}
	if strings.HasPrefix(lower, "grpcs://") {
		return parsedGRPCEndpoint{Target: strings.TrimSpace(e[len("grpcs://"):]), ServerTLSScheme: true}
	}
	if strings.HasPrefix(lower, "http://") {
		return parsedGRPCEndpoint{Target: strings.TrimSpace(e[len("http://"):]), ServerTLSScheme: false}
	}
	if strings.HasPrefix(lower, "grpc://") {
		return parsedGRPCEndpoint{Target: strings.TrimSpace(e[len("grpc://"):]), ServerTLSScheme: false}
	}
	return parsedGRPCEndpoint{Target: e, ServerTLSScheme: false}
}

func stripBearerPrefix(token string) string {
	t := strings.TrimSpace(token)
	if len(t) >= 7 && strings.EqualFold(t[:7], "bearer ") {
		return strings.TrimSpace(t[7:])
	}
	return t
}

func serverNameForDial(target string) string {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return target
	}
	return host
}

func dialOptions(p Params) ([]grpc.DialOption, bool, error) {
	parsed := parseGRPCEndpoint(p.Endpoint)
	var opts []grpc.DialOption
	tlsInUse := false

	switch {
	case p.CertPath != "" && p.CertKeyPath != "":
		cert, err := tls.LoadX509KeyPair(p.CertPath, p.CertKeyPath)
		if err != nil {
			return nil, false, fmt.Errorf("load client certificate: %w", err)
		}
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			ServerName:   serverNameForDial(parsed.Target),
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
		tlsInUse = true
	case p.CertPath != "" || p.CertKeyPath != "":
		return nil, false, fmt.Errorf("both cert_path and cert_key_path must be set together for TLS client authentication")
	case parsed.ServerTLSScheme:
		tlsCfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverNameForDial(parsed.Target),
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
		tlsInUse = true
	default:
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	token := stripBearerPrefix(p.Token)
	if token != "" {
		if !tlsInUse {
			return nil, false, fmt.Errorf("bearer token requires TLS: use https:// or grpcs:// in permify_url, or set cert_path and cert_key_path for mTLS")
		}
		md := map[string]string{"authorization": "Bearer " + token}
		opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials(md)))
	}

	return opts, tlsInUse, nil
}

func New(p Params) (*permify.Client, error) {
	target := parseGRPCEndpoint(p.Endpoint).Target
	if target == "" {
		return nil, fmt.Errorf("endpoint is empty")
	}
	opts, _, err := dialOptions(p)
	if err != nil {
		return nil, err
	}
	client, err := permify.NewClient(
		permify.Config{
			Endpoint: target,
		},
		opts...,
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
