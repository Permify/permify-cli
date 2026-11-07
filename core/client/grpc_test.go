package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestNewClientCredentialValidation(t *testing.T) {
	config.CliConfig = config.CoreConfig{}

	_, err := New("")
	if err == nil {
		t.Fatalf("expected error when both endpoint argument and stored PermifyURL are empty")
	}

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		CertKey:    "/nonexistent/key.pem",
	}
	_, err = New("")
	if err == nil {
		t.Fatalf("expected error when cert_key is set without cert_path")
	}

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		CertPath:   "/nonexistent/cert.pem",
	}
	_, err = New("")
	if err == nil {
		t.Fatalf("expected error when cert_path points to missing certificate file")
	}
}

func TestConfigCredentialStorageRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		Tenant:     "t1",
		Token:      "secret-token-123",
		CertPath:   "/etc/permify/tls.crt",
		CertKey:    "/etc/permify/tls.key",
	}
	if err := config.New(cfgPath, "default"); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config.CliConfig = config.CoreConfig{}
	if err := config.Load(cfgPath, "default"); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.CliConfig.PermifyURL != "localhost:3478" ||
		config.CliConfig.Token != "secret-token-123" ||
		config.CliConfig.CertPath != "/etc/permify/tls.crt" ||
		config.CliConfig.CertKey != "/etc/permify/tls.key" {
		t.Fatalf("unexpected loaded credentials: %+v", config.CliConfig)
	}
	_ = os.Remove(cfgPath)
}
