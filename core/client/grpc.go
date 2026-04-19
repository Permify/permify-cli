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

func normalizeGRPCEndpoint(endpoint string) string {
	e := strings.TrimSpace(endpoint)
	lower := strings.ToLower(e)
	for _, prefix := range []string{"https://", "http://", "grpc://", "grpcs://"} {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(e[len(prefix):])
		}
	}
	return e
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
	var opts []grpc.DialOption
	tlsInUse := false

	switch {
	case p.CertPath != "" && p.CertKeyPath != "":
		cert, err := tls.LoadX509KeyPair(p.CertPath, p.CertKeyPath)
		if err != nil {
			return nil, false, fmt.Errorf("load client certificate: %w", err)
		}
		target := normalizeGRPCEndpoint(p.Endpoint)
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			ServerName:   serverNameForDial(target),
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
		tlsInUse = true
	case p.CertPath != "" || p.CertKeyPath != "":
		return nil, false, fmt.Errorf("both cert_path and cert_key_path must be set together for TLS client authentication")
	default:
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	token := stripBearerPrefix(p.Token)
	if token != "" {
		md := map[string]string{"authorization": "Bearer " + token}
		if tlsInUse {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials(md)))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(md)))
		}
	}

	return opts, tlsInUse, nil
}

func New(p Params) (*permify.Client, error) {
	target := normalizeGRPCEndpoint(p.Endpoint)
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
