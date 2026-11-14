package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreAndRetrieveCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	certFile := filepath.Join(tmpDir, "server.crt")
	keyFile := filepath.Join(tmpDir, "server.key")

	if err := os.WriteFile(certFile, []byte("dummy-cert"), 0600); err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("dummy-key"), 0600); err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}

	CliConfig = CoreConfig{
		PermifyURL: "https://localhost:3478",
		Tenant:     "t1",
		Token:      "secret-token-123",
		CertPath:   certFile,
		CertKey:    keyFile,
	}

	if err := New(configFile, "default"); err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	// Reset global config and reload from disk
	CliConfig = CoreConfig{}
	if err := IsConfigured(configFile, "default"); err != nil {
		t.Fatalf("IsConfigured() returned unexpected error: %v", err)
	}
	if err := Load(configFile, "default"); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	creds, err := GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials() returned unexpected error: %v", err)
	}

	if creds.PermifyURL != "https://localhost:3478" {
		t.Errorf("expected PermifyURL https://localhost:3478, got %s", creds.PermifyURL)
	}
	if creds.Tenant != "t1" {
		t.Errorf("expected Tenant t1, got %s", creds.Tenant)
	}
	if creds.Token != "secret-token-123" {
		t.Errorf("expected Token secret-token-123, got %s", creds.Token)
	}
	if creds.CertPath != certFile {
		t.Errorf("expected CertPath %s, got %s", certFile, creds.CertPath)
	}
	if creds.CertKey != keyFile {
		t.Errorf("expected CertKey %s, got %s", keyFile, creds.CertKey)
	}
	if !creds.SslEnabled {
		t.Errorf("expected SslEnabled true")
	}
}

func TestValidateCredentialsErrors(t *testing.T) {
	cfg := CoreConfig{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty PermifyURL, got nil")
	}

	cfg = CoreConfig{
		PermifyURL: "localhost:3478",
		CertKey:    "/nonexistent/key.pem",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when CertKey is set without CertPath, got nil")
	}

	cfg = CoreConfig{
		PermifyURL: "localhost:3478",
		CertPath:   "/nonexistent/cert.pem",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for nonexistent CertPath, got nil")
	}
}
