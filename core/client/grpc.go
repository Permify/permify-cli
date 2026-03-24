// Package client handles the permify client to connect with the server
package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// New initializes a new permify client
func New(endpoint string) (*permify.Client, error) {
	if endpoint == "" {
		endpoint = config.CliConfig.PermifyURL
	}
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	opts := []grpc.DialOption{}
	transportSecure := false

	certPath := strings.TrimSpace(config.CliConfig.CertPath)
	certKey := strings.TrimSpace(config.CliConfig.CertKey)

	switch {
	case certPath != "" && certKey != "":
		certificate, err := tls.LoadX509KeyPair(certPath, certKey)
		if err != nil {
			return nil, err
		}
		caCert, err := os.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to parse certificate at %s", certPath)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})))
		transportSecure = true
	case certPath != "":
		tlsCreds, err := credentials.NewClientTLSFromFile(certPath, "")
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
		transportSecure = true
	case certKey != "":
		return nil, fmt.Errorf("cert_key provided without cert_path")
	default:
		// Todo: Implement secure call with tls certificate
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	token := strings.TrimSpace(config.CliConfig.Token)
	if token != "" {
		metadata := map[string]string{"authorization": token}
		if transportSecure {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials(metadata)))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(metadata)))
		}
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: endpoint,
		},
		opts...,
	)
	return client, err
}
