package client

import (
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestDialOptionsAllowStoredTokenWithoutTLS(t *testing.T) {
	opts, err := dialOptions(config.CoreConfig{
		Token: "secret-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 2 {
		t.Fatalf("expected transport and token options, got %d", len(opts))
	}
}

func TestDialOptionsAllowStoredTokenWithTLS(t *testing.T) {
	opts, err := dialOptions(config.CoreConfig{
		SslEnabled: true,
		Token:      "secret-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 2 {
		t.Fatalf("expected TLS transport and token options, got %d", len(opts))
	}
}

func TestDialOptionsRejectCertKeyWithoutCert(t *testing.T) {
	_, err := dialOptions(config.CoreConfig{
		CertKeyPath: "/tmp/client.key",
	})
	if err == nil {
		t.Fatal("expected cert_key_path without cert_path to fail")
	}
}
