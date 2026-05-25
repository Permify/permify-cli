package client

import (
	"context"
	"testing"

	"github.com/Permify/permify-cli/core/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestDialOptionsUseInsecureCredentialsForHTTP(t *testing.T) {
	opts, err := DialOptions(config.CoreConfig{PermifyURL: "http://localhost:3478"})
	if err != nil {
		t.Fatalf("DialOptions returned error: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("expected only transport credentials option, got %d", len(opts))
	}
}

func TestDialOptionsAddBearerTokenPerRPCCredentials(t *testing.T) {
	opts, err := DialOptions(config.CoreConfig{PermifyURL: "http://localhost:3478", Token: "secret-token"})
	if err != nil {
		t.Fatalf("DialOptions returned error: %v", err)
	}
	if len(opts) != 2 {
		t.Fatalf("expected transport credentials and token credentials options, got %d", len(opts))
	}

	creds := bearerTokenCredentials{token: "secret-token", requireTransportSecurity: false}
	metadata, err := creds.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata returned error: %v", err)
	}
	if metadata["authorization"] != "Bearer secret-token" {
		t.Fatalf("expected bearer token authorization metadata, got %q", metadata["authorization"])
	}
}

func TestNewFromConfigUsesConfiguredEndpoint(t *testing.T) {
	factory := func(c permifyConfig, opts ...grpc.DialOption) (permifyClient, error) {
		if c.Endpoint != "localhost:3478" {
			t.Fatalf("expected configured endpoint, got %q", c.Endpoint)
		}
		if len(opts) != 1 {
			t.Fatalf("expected generated dial options to be passed to client factory")
		}
		return nil, nil
	}

	_, err := newFromConfig(config.CoreConfig{PermifyURL: "localhost:3478"}, factory)
	if err != nil {
		t.Fatalf("newFromConfig returned error: %v", err)
	}
}

var _ grpc.DialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
