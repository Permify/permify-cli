package client

import (
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestIsSecureEndpoint(t *testing.T) {
	tests := map[string]bool{
		"https://permify.example.com": true,
		"grpcs://permify.example.com": true,
		"http://permify.example.com":  false,
		"localhost:3478":              false,
	}

	for endpoint, want := range tests {
		if got := isSecureEndpoint(endpoint); got != want {
			t.Fatalf("isSecureEndpoint(%q) = %v, want %v", endpoint, got, want)
		}
	}
}

func TestTransportCredentialsRequiresCertPair(t *testing.T) {
	_, err := transportCredentials("localhost:3478", config.CoreConfig{
		CertPath: "client.crt",
	})
	if err == nil {
		t.Fatal("expected error when only cert path is configured")
	}
}

func TestDialOptionsAllowsStoredToken(t *testing.T) {
	opts, err := dialOptions("localhost:3478", config.CoreConfig{
		Token: "test-token",
	})
	if err != nil {
		t.Fatalf("dialOptions returned error: %v", err)
	}
	if len(opts) != 2 {
		t.Fatalf("dialOptions returned %d options, want 2", len(opts))
	}
}
