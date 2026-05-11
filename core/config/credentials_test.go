package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCredentialsFilePathForHome(t *testing.T) {
	got := credentialsFilePathForHome("/tmp/permify-home")
	want := filepath.Join("/tmp/permify-home", ".permify", "credentials")
	if got != want {
		t.Fatalf("credentialsFilePathForHome() = %q, want %q", got, want)
	}
}

func TestWriteAndLoadStoredCredentials(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := CoreConfig{
		PermifyURL:  "https://permify.example:3478",
		Token:       "secret-token",
		CertPath:    "/tmp/client.crt",
		CertKeyPath: "/tmp/client.key",
	}

	if err := WriteStoredCredentials("default", cfg); err != nil {
		t.Fatalf("WriteStoredCredentials() error = %v", err)
	}

	path, err := credentialsFilePath()
	if err != nil {
		t.Fatalf("credentialsFilePath() error = %v", err)
	}

	stored, err := LoadStoredCredentials("default")
	if err != nil {
		t.Fatalf("LoadStoredCredentials() error = %v", err)
	}

	if stored.PermifyURL != cfg.PermifyURL {
		t.Fatalf("stored endpoint = %q, want %q", stored.PermifyURL, cfg.PermifyURL)
	}
	if stored.Token != cfg.Token {
		t.Fatalf("stored token = %q, want %q", stored.Token, cfg.Token)
	}
	if stored.CertPath != cfg.CertPath {
		t.Fatalf("stored cert path = %q, want %q", stored.CertPath, cfg.CertPath)
	}
	if stored.CertKeyPath != cfg.CertKeyPath {
		t.Fatalf("stored cert key path = %q, want %q", stored.CertKeyPath, cfg.CertKeyPath)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("os.Stat() error = %v", err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Fatalf("credentials file mode = %04o, want 0600", mode)
		}
	}
}

func TestLoadPrefersStoredCredentialEndpoint(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	oldCliConfig := CliConfig
	oldProfileConfigs := profileConfigs
	defer func() {
		CliConfig = oldCliConfig
		profileConfigs = oldProfileConfigs
	}()

	configPath := filepath.Join(home, "permctl.yaml")
	configContents := []byte("default:\n  permify_url: localhost:3478\n  tenant: tenant-1\n")
	if err := os.WriteFile(configPath, configContents, 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	if err := WriteStoredCredentials("default", CoreConfig{
		PermifyURL:  "https://secure.permify.example:3478",
		Token:       "stored-token",
		CertPath:    "/tmp/client.crt",
		CertKeyPath: "/tmp/client.key",
	}); err != nil {
		t.Fatalf("WriteStoredCredentials() error = %v", err)
	}

	if err := Load(configPath, "default"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if CliConfig.PermifyURL != "https://secure.permify.example:3478" {
		t.Fatalf("CliConfig.PermifyURL = %q, want stored endpoint", CliConfig.PermifyURL)
	}
	if CliConfig.Tenant != "tenant-1" {
		t.Fatalf("CliConfig.Tenant = %q, want tenant-1", CliConfig.Tenant)
	}
	if CliConfig.Token != "stored-token" {
		t.Fatalf("CliConfig.Token = %q, want stored-token", CliConfig.Token)
	}
}
