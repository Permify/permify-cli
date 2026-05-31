package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndLoadPreservesStoredCredentials(t *testing.T) {
	t.Cleanup(func() {
		CliConfig = CoreConfig{}
		profileConfigs = ProfileConfigs{}
	})

	configFile := filepath.Join(t.TempDir(), "permctl.yaml")
	profile := "work"

	CliConfig = CoreConfig{
		PermifyURL:  "permify.example:3478",
		Tenant:      "tenant-1",
		Token:       "secret-token",
		CertPath:    "certs/client.pem",
		CertKeyPath: "certs/client.key",
	}

	if err := New(configFile, profile); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	CliConfig = CoreConfig{
		PermifyURL:  "permify.example:9443",
		Tenant:      "tenant-2",
		Token:       "updated-token",
		CertPath:    "certs/updated.pem",
		CertKeyPath: "certs/updated.key",
	}

	if err := Write(); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}

	if err := Load(configFile, profile); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := CliConfig.PermifyURL, "permify.example:9443"; got != want {
		t.Fatalf("PermifyURL = %q, want %q", got, want)
	}
	if got, want := CliConfig.Tenant, "tenant-2"; got != want {
		t.Fatalf("Tenant = %q, want %q", got, want)
	}
	if got, want := CliConfig.Token, "updated-token"; got != want {
		t.Fatalf("Token = %q, want %q", got, want)
	}
	if got, want := CliConfig.CertPath, "certs/updated.pem"; got != want {
		t.Fatalf("CertPath = %q, want %q", got, want)
	}
	if got, want := CliConfig.CertKeyPath, "certs/updated.key"; got != want {
		t.Fatalf("CertKeyPath = %q, want %q", got, want)
	}
	if !CliConfig.SslEnabled {
		t.Fatal("SslEnabled = false, want true when certificate path is configured")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected config file to be written")
	}
}
