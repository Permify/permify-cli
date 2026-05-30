package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteAndLoadCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	credentialFile := filepath.Join(tmpDir, "credentials")

	originalCredentials := CliCredentials
	originalCredentialProfiles := credentialsConfigs
	t.Cleanup(func() {
		CliCredentials = originalCredentials
		credentialsConfigs = originalCredentialProfiles
	})

	credentialsConfigs = CredentialProfiles{
		Configs: map[string]ConnectionCredentials{},
		File:    credentialFile,
		Profile: "default",
	}
	CliCredentials = ConnectionCredentials{
		Endpoint: "localhost:3478",
		Token:    "secret-token",
		CertPath: "/tmp/client.pem",
		CertKey:  "/tmp/client-key.pem",
	}

	if err := WriteCredentials(); err != nil {
		t.Fatalf("WriteCredentials() error = %v", err)
	}

	info, err := os.Stat(credentialFile)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected file mode: got %v want 0600", info.Mode().Perm())
	}

	CliCredentials = ConnectionCredentials{}
	credentialsConfigs = CredentialProfiles{
		Configs: map[string]ConnectionCredentials{},
		File:    credentialFile,
		Profile: "default",
	}

	if err := LoadCredentials(credentialFile, "default"); err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if CliCredentials.Endpoint != "localhost:3478" {
		t.Fatalf("unexpected endpoint: got %q", CliCredentials.Endpoint)
	}
	if CliCredentials.Token != "secret-token" {
		t.Fatalf("unexpected token: got %q", CliCredentials.Token)
	}
	if CliCredentials.CertPath != "/tmp/client.pem" {
		t.Fatalf("unexpected cert path: got %q", CliCredentials.CertPath)
	}
	if CliCredentials.CertKey != "/tmp/client-key.pem" {
		t.Fatalf("unexpected cert key: got %q", CliCredentials.CertKey)
	}
}
