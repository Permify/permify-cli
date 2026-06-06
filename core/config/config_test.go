package config

import (
	"os"
	"testing"
)

func TestConfigPersistsConnectionCredentials(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "permctl-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	CliConfig = CoreConfig{
		PermifyURL:  "https://permify.example.com:3478",
		Tenant:      "tenant-1",
		Token:       "secret-token",
		CertPath:    "/tmp/client.crt",
		CertKeyPath: "/tmp/client.key",
	}

	if err := New(file.Name(), "default"); err != nil {
		t.Fatal(err)
	}

	CliConfig = CoreConfig{}

	if err := Load(file.Name(), "default"); err != nil {
		t.Fatal(err)
	}

	if CliConfig.PermifyURL != "https://permify.example.com:3478" {
		t.Fatalf("unexpected endpoint: %q", CliConfig.PermifyURL)
	}
	if CliConfig.Tenant != "tenant-1" {
		t.Fatalf("unexpected tenant: %q", CliConfig.Tenant)
	}
	if CliConfig.Token != "secret-token" {
		t.Fatalf("unexpected token: %q", CliConfig.Token)
	}
	if CliConfig.CertPath != "/tmp/client.crt" {
		t.Fatalf("unexpected cert path: %q", CliConfig.CertPath)
	}
	if CliConfig.CertKeyPath != "/tmp/client.key" {
		t.Fatalf("unexpected cert key path: %q", CliConfig.CertKeyPath)
	}
	if !CliConfig.SslEnabled {
		t.Fatal("expected https endpoint to enable TLS")
	}
}
