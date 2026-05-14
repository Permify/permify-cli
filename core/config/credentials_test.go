package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolateConfigTest(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}
	return tmp
}

func TestWriteStoresCredentialsSeparately(t *testing.T) {
	tmp := isolateConfigTest(t)
	configFile := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(configFile, []byte("dev:\n  tenant: old\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	CliConfig = CoreConfig{
		PermifyURL:  "https://permify.example:3478",
		Tenant:      "tenant-1",
		Token:       "secret-token",
		CertPath:    "client.crt",
		CertKeyPath: "client.key",
	}
	profileConfigs = ProfileConfigs{
		File:    configFile,
		Profile: "dev",
		Configs: map[string]CoreConfig{"dev": CliConfig},
	}

	if err := Write(); err != nil {
		t.Fatalf("write: %v", err)
	}

	configData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configText := string(configData)
	for _, unexpected := range []string{"permify_url", "secret-token", "client.crt", "client.key"} {
		if strings.Contains(configText, unexpected) {
			t.Fatalf("config file contains credential data %q:\n%s", unexpected, configText)
		}
	}
	if !strings.Contains(configText, "tenant-1") {
		t.Fatalf("config file should retain tenant:\n%s", configText)
	}

	credentialsData, err := os.ReadFile(filepath.Join(tmp, ".permify", "credentials"))
	if err != nil {
		t.Fatalf("read credentials: %v", err)
	}
	credentialsText := string(credentialsData)
	for _, expected := range []string{"permify_url", "https://permify.example:3478", "secret-token", "client.crt", "client.key"} {
		if !strings.Contains(credentialsText, expected) {
			t.Fatalf("credentials file missing %q:\n%s", expected, credentialsText)
		}
	}
}

func TestLoadMergesStoredCredentials(t *testing.T) {
	tmp := isolateConfigTest(t)
	configFile := filepath.Join(tmp, "config.yaml")
	credentialsFile := filepath.Join(tmp, ".permify", "credentials")
	if err := os.WriteFile(configFile, []byte("dev:\n  tenant: tenant-1\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(credentialsFile), 0700); err != nil {
		t.Fatalf("mkdir credentials: %v", err)
	}
	credentials := []byte("dev:\n  permify_url: https://permify.example:3478\n  token: secret-token\n  cert_path: client.crt\n  cert_key_path: client.key\n")
	if err := os.WriteFile(credentialsFile, credentials, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	if err := Load(configFile, "dev"); err != nil {
		t.Fatalf("load: %v", err)
	}
	if CliConfig.PermifyURL != "https://permify.example:3478" {
		t.Fatalf("PermifyURL=%q", CliConfig.PermifyURL)
	}
	if CliConfig.Tenant != "tenant-1" {
		t.Fatalf("Tenant=%q", CliConfig.Tenant)
	}
	if CliConfig.Token != "secret-token" {
		t.Fatalf("Token=%q", CliConfig.Token)
	}
	if CliConfig.CertPath != "client.crt" || CliConfig.CertKeyPath != "client.key" {
		t.Fatalf("cert paths=%q/%q", CliConfig.CertPath, CliConfig.CertKeyPath)
	}
	if !CliConfig.SslEnabled {
		t.Fatalf("expected https endpoint to enable ssl")
	}
}

func TestIsConfiguredUsesStoredEndpoint(t *testing.T) {
	tmp := isolateConfigTest(t)
	configFile := filepath.Join(tmp, "config.yaml")
	credentialsFile := filepath.Join(tmp, ".permify", "credentials")
	if err := os.WriteFile(configFile, []byte("dev:\n  tenant: tenant-1\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(credentialsFile), 0700); err != nil {
		t.Fatalf("mkdir credentials: %v", err)
	}
	if err := os.WriteFile(credentialsFile, []byte("dev:\n  permify_url: localhost:3478\n"), 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	if err := IsConfigured(configFile, "dev"); err != nil {
		t.Fatalf("expected configured profile: %v", err)
	}
}
