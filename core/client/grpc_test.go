package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestResolveDialOptions_MissingEndpoint(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	_, _, err := ResolveDialOptions("", "", "", "")
	if err == nil {
		t.Fatal("expected error when endpoint is missing, got nil")
	}
}

func TestResolveDialOptions_InsecureWithToken(t *testing.T) {
	endpoint, opts, err := ResolveDialOptions("localhost:3478", "secret-token", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != "localhost:3478" {
		t.Fatalf("expected endpoint localhost:3478, got %s", endpoint)
	}
	if len(opts) != 2 {
		t.Fatalf("expected 2 dial options (transport + token credentials), got %d", len(opts))
	}
}

func TestResolveDialOptions_MismatchedCertAndKey(t *testing.T) {
	_, _, err := ResolveDialOptions("localhost:3478", "", "/tmp/cert.pem", "")
	if err == nil {
		t.Fatal("expected error when cert_key is missing, got nil")
	}
}

func TestConfigCredentialStorageRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")

	config.CliConfig = config.CoreConfig{
		PermifyURL: "https://permify.example.com:3478",
		Tenant:     "t1",
		Token:      "tok_live_123",
		CertPath:   "/etc/permify/tls.crt",
		CertKey:    "/etc/permify/tls.key",
	}

	if err := config.New(cfgFile, "default"); err != nil {
		t.Fatalf("config.New failed: %v", err)
	}

	config.CliConfig = config.CoreConfig{}
	if err := config.Load(cfgFile, "default"); err != nil {
		t.Fatalf("config.Load failed: %v", err)
	}

	endpoint, token, certPath, certKey := config.GetCredentials()
	if endpoint != "https://permify.example.com:3478" || token != "tok_live_123" || certPath != "/etc/permify/tls.crt" || certKey != "/etc/permify/tls.key" {
		t.Fatalf("unexpected credentials roundtrip: endpoint=%q token=%q certPath=%q certKey=%q", endpoint, token, certPath, certKey)
	}
	if !config.CliConfig.SslEnabled {
		t.Fatal("expected SslEnabled=true when https and TLS cert paths are configured")
	}
	_ = os.Remove(cfgFile)
}
