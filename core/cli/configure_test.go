package cli

import (
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestBuildConfigPreservesCredentialFields(t *testing.T) {
	cfg := buildConfig("https://permify.example.com", "tenant-1", "token-1", "/tmp/client.crt", "/tmp/client.key")

	if cfg.PermifyURL != "https://permify.example.com" {
		t.Fatalf("expected permify url to be stored, got %q", cfg.PermifyURL)
	}
	if cfg.Tenant != "tenant-1" {
		t.Fatalf("expected tenant to be stored, got %q", cfg.Tenant)
	}
	if cfg.Token != "token-1" {
		t.Fatalf("expected token to be stored, got %q", cfg.Token)
	}
	if cfg.CertPath != "/tmp/client.crt" {
		t.Fatalf("expected cert path to be stored, got %q", cfg.CertPath)
	}
	if cfg.CertKey != "/tmp/client.key" {
		t.Fatalf("expected cert key to be stored, got %q", cfg.CertKey)
	}
}

var _ config.CoreConfig = buildConfig("", "", "", "", "")
