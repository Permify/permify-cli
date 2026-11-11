package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreAndRetrieveCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yaml")
	certFile := filepath.Join(tmpDir, "client.crt")
	keyFile := filepath.Join(tmpDir, "client.key")

	if err := os.WriteFile(certFile, []byte("dummy-cert"), 0600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("dummy-key"), 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	CliConfig = CoreConfig{
		PermifyURL: "https://api.permify.example:3478",
		Tenant:     "t1",
		Token:      "secret-token",
		CertPath:   certFile,
		CertKey:    keyFile,
	}

	if err := New(cfgFile, "default"); err != nil {
		t.Fatalf("New failed: %v", err)
	}

	CliConfig = CoreConfig{}
	if err := Load(cfgFile, "default"); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	creds, err := GetStoredCredentials("")
	if err != nil {
		t.Fatalf("GetStoredCredentials failed: %v", err)
	}
	if creds.PermifyURL != "https://api.permify.example:3478" || creds.Token != "secret-token" || creds.CertPath != certFile || creds.CertKey != keyFile {
		t.Fatalf("unexpected credentials loaded: %+v", creds)
	}
}

func TestGetStoredCredentialsValidationErrors(t *testing.T) {
	CliConfig = CoreConfig{}
	if _, err := GetStoredCredentials(""); err == nil {
		t.Fatal("expected error when endpoint is missing")
	}

	CliConfig = CoreConfig{
		PermifyURL: "localhost:3478",
		CertKey:    "/nonexistent/key.pem",
	}
	if _, err := GetStoredCredentials(""); err == nil {
		t.Fatal("expected error when cert_key is set without cert_path")
	}
}
