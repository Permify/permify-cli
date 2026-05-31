package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteStoresCredentialsInDedicatedFile(t *testing.T) {
	t.Cleanup(resetConfigGlobals)
	configFile, credentialsFile := configureTestPaths(t)
	profile := "work"

	CliConfig = CoreConfig{
		PermifyURL:  "permify.example:9443",
		Tenant:      "tenant-2",
		Token:       "updated-token",
		CertPath:    "certs/updated.pem",
		CertKeyPath: "certs/updated.key",
	}

	if err := New(configFile, profile); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := Write(); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	configData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	configText := string(configData)
	for _, forbidden := range []string{"permify.example:9443", "updated-token", "certs/updated.pem", "certs/updated.key"} {
		if strings.Contains(configText, forbidden) {
			t.Fatalf("primary config unexpectedly contains %q", forbidden)
		}
	}
	if !strings.Contains(configText, "tenant: tenant-2") {
		t.Fatalf("primary config missing tenant entry: %s", configText)
	}

	credentialsData, err := os.ReadFile(credentialsFile)
	if err != nil {
		t.Fatalf("ReadFile(credentials) error = %v", err)
	}
	credentialsText := string(credentialsData)
	for _, expected := range []string{"permify.example:9443", "updated-token", "certs/updated.pem", "certs/updated.key"} {
		if !strings.Contains(credentialsText, expected) {
			t.Fatalf("credentials file missing %q", expected)
		}
	}
}

func TestLoadMergesPrimaryConfigWithStoredCredentials(t *testing.T) {
	t.Cleanup(resetConfigGlobals)
	configFile, _ := configureTestPaths(t)
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

	if err := Write(); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}

	if err := Load(configFile, profile); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := CliConfig.PermifyURL, "permify.example:3478"; got != want {
		t.Fatalf("PermifyURL = %q, want %q", got, want)
	}
	if got, want := CliConfig.Tenant, "tenant-1"; got != want {
		t.Fatalf("Tenant = %q, want %q", got, want)
	}
	if got, want := CliConfig.Token, "secret-token"; got != want {
		t.Fatalf("Token = %q, want %q", got, want)
	}
	if got, want := CliConfig.CertPath, "certs/client.pem"; got != want {
		t.Fatalf("CertPath = %q, want %q", got, want)
	}
	if got, want := CliConfig.CertKeyPath, "certs/client.key"; got != want {
		t.Fatalf("CertKeyPath = %q, want %q", got, want)
	}
	if !CliConfig.SslEnabled {
		t.Fatal("SslEnabled = false, want true when certificate path is configured")
	}
}

func TestIsConfiguredAcceptsEndpointFromCredentialsStore(t *testing.T) {
	t.Cleanup(resetConfigGlobals)
	configFile, _ := configureTestPaths(t)
	profile := "work"

	CliConfig = CoreConfig{
		PermifyURL: "permify.example:3478",
		Tenant:     "tenant-1",
	}

	if err := New(configFile, profile); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := Write(); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := IsConfigured(configFile, profile); err != nil {
		t.Fatalf("IsConfigured() error = %v", err)
	}
}

func configureTestPaths(t *testing.T) (string, string) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	configFile := filepath.Join(home, "permctl.yaml")
	credentialsFile := filepath.Join(home, ".permify", "credentials")
	return configFile, credentialsFile
}

func resetConfigGlobals() {
	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}
}
