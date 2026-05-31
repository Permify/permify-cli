package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigNewWriteLoad_PreservesFieldsAndRestrictsPermissions(t *testing.T) {
	oldCliConfig := CliConfig
	oldProfileConfigs := profileConfigs
	t.Cleanup(func() {
		CliConfig = oldCliConfig
		profileConfigs = oldProfileConfigs
	})

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "permctl.yaml")

	CliConfig = CoreConfig{
		PermifyURL: "localhost:3478",
		Token:      "my-token",
		CertPath:   "/tmp/ca.pem",
		CertKey:    "/tmp/client.key",
		Tenant:     "tenant-id",
	}

	if err := New(configFile, "default"); err != nil {
		t.Fatalf("New() error: %v", err)
	}

	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected config perms 0600, got %v", got)
	}

	if err := Load(configFile, "default"); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if CliConfig.PermifyURL != "localhost:3478" {
		t.Fatalf("PermifyURL mismatch: %q", CliConfig.PermifyURL)
	}
	if CliConfig.Token != "my-token" {
		t.Fatalf("Token mismatch: %q", CliConfig.Token)
	}
	if CliConfig.CertPath != "/tmp/ca.pem" {
		t.Fatalf("CertPath mismatch: %q", CliConfig.CertPath)
	}
	if CliConfig.CertKey != "/tmp/client.key" {
		t.Fatalf("CertKey mismatch: %q", CliConfig.CertKey)
	}
	if CliConfig.Tenant != "tenant-id" {
		t.Fatalf("Tenant mismatch: %q", CliConfig.Tenant)
	}
	if !CliConfig.SslEnabled {
		t.Fatalf("expected SslEnabled to be true when cert material is configured")
	}
}
