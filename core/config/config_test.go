package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPersistAndLoadCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	certFile := filepath.Join(tmpDir, "client.crt")
	keyFile := filepath.Join(tmpDir, "client.key")

	if err := os.WriteFile(certFile, []byte("cert-data"), 0600); err != nil {
		t.Fatalf("failed to create cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("key-data"), 0600); err != nil {
		t.Fatalf("failed to create key file: %v", err)
	}

	CliConfig = CoreConfig{
		PermifyURL: "grpcs://api.permify.co:3478",
		Tenant:     "t1",
		Token:      "secret-token-123",
		CertPath:   certFile,
		CertKey:    keyFile,
	}

	if err := New(cfgFile, "default"); err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	info, err := os.Stat(cfgFile)
	if err != nil {
		t.Fatalf("Stat() failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}

	CliConfig = CoreConfig{}
	if err := Load(cfgFile, "default"); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if CliConfig.PermifyURL != "grpcs://api.permify.co:3478" {
		t.Errorf("unexpected PermifyURL: %s", CliConfig.PermifyURL)
	}
	if CliConfig.Tenant != "t1" {
		t.Errorf("unexpected Tenant: %s", CliConfig.Tenant)
	}
	if CliConfig.Token != "secret-token-123" {
		t.Errorf("unexpected Token: %s", CliConfig.Token)
	}
	if CliConfig.CertPath != certFile {
		t.Errorf("unexpected CertPath: %s", CliConfig.CertPath)
	}
	if CliConfig.CertKey != keyFile {
		t.Errorf("unexpected CertKey: %s", CliConfig.CertKey)
	}
	if !CliConfig.SslEnabled {
		t.Errorf("expected SslEnabled to be true")
	}
	if err := CliConfig.Validate(); err != nil {
		t.Errorf("expected Validate() to succeed, got: %v", err)
	}
}

func TestConfigValidateErrors(t *testing.T) {
	cfg := CoreConfig{}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error for empty endpoint")
	}

	cfg = CoreConfig{
		PermifyURL: "localhost:3478",
		CertPath:   "/some/cert.pem",
	}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error when cert_key is missing")
	}

	cfg = CoreConfig{
		PermifyURL: "localhost:3478",
		CertPath:   "/nonexistent/cert.pem",
		CertKey:    "/nonexistent/key.pem",
	}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error when certificate files do not exist")
	}
}
