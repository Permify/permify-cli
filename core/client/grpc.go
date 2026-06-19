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
	clientConfig := config.CliConfig
	if endpoint != "" {
		clientConfig.PermifyURL = endpoint
		clientConfig.SslEnabled = strings.HasPrefix(endpoint, "https")
	}

	options := []grpc.DialOption{}
	if clientConfig.SslEnabled {
		transportCredentials, err := tlsCredentials(clientConfig)
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.WithTransportCredentials(transportCredentials))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if clientConfig.Token != "" {
		token := map[string]string{"authorization": fmt.Sprintf("Bearer %s", clientConfig.Token)}
		if clientConfig.SslEnabled {
			options = append(options, grpc.WithPerRPCCredentials(secureTokenCredentials(token)))
		} else {
			options = append(options, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(token)))
		}
	}

	client, err := permify.NewClient(
		permify.Config{
			Endpoint: clientConfig.PermifyURL,
		},
		options...,
	)
	return client, err
}

func tlsCredentials(clientConfig config.CoreConfig) (credentials.TransportCredentials, error) {
	tlsConfig := &tls.Config{}
	if clientConfig.CertificatePath == "" && clientConfig.CertificateKeyPath == "" {
		return credentials.NewTLS(tlsConfig), nil
	}
	if clientConfig.CertificatePath == "" || clientConfig.CertificateKeyPath == "" {
		return nil, fmt.Errorf("both certificate_path and certificate_key_path are required when configuring client certificates")
	}
	cert, err := tls.LoadX509KeyPair(clientConfig.CertificatePath, clientConfig.CertificateKeyPath)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}
	return credentials.NewTLS(tlsConfig), nil
}
