package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func withCredentialsPath(t *testing.T) string {
	t.Helper()

	oldPath := credentialsPathOverride
	oldCliConfig := CliConfig
	oldProfileConfigs := profileConfigs
	path := filepath.Join(t.TempDir(), ".permify", "credentials")
	credentialsPathOverride = path
	t.Cleanup(func() {
		credentialsPathOverride = oldPath
		CliConfig = oldCliConfig
		profileConfigs = oldProfileConfigs
	})
	return path
}

func TestCredentialsFilePathForHome(t *testing.T) {
	home := filepath.Join("tmp", "permify-home")
	got := credentialsFilePathForHome(home)
	want := filepath.Join(home, ".permify", "credentials")
	if got != want {
		t.Fatalf("credentialsFilePathForHome() = %q, want %q", got, want)
	}
}

func TestWriteAndLoadStoredCredentialsByProfile(t *testing.T) {
	path := withCredentialsPath(t)

	defaultConfig := CoreConfig{
		PermifyURL: "https://permify.example:3478",
		Token:      "secret-token",
		CertPath:  "/tmp/client.crt",
		CertKey:   "/tmp/client.key",
	}
	stagingConfig := CoreConfig{
		PermifyURL: "grpc://staging.example:3478",
		Token:      "staging-token",
	}

	if err := WriteStoredCredentials("default", defaultConfig); err != nil {
		t.Fatalf("WriteStoredCredentials(default) error = %v", err)
	}
	if err := WriteStoredCredentials("staging", stagingConfig); err != nil {
		t.Fatalf("WriteStoredCredentials(staging) error = %v", err)
	}

	stored, err := LoadStoredCredentials("default")
	if err != nil {
		t.Fatalf("LoadStoredCredentials(default) error = %v", err)
	}
	if stored.EndpointValue() != defaultConfig.PermifyURL {
		t.Fatalf("default endpoint = %q, want %q", stored.EndpointValue(), defaultConfig.PermifyURL)
	}
	if stored.Token != defaultConfig.Token {
		t.Fatalf("default token = %q, want %q", stored.Token, defaultConfig.Token)
	}
	if stored.CertPath != defaultConfig.CertPath {
		t.Fatalf("default cert path = %q, want %q", stored.CertPath, defaultConfig.CertPath)
	}
	if stored.CertKeyValue() != defaultConfig.CertKey {
		t.Fatalf("default cert key = %q, want %q", stored.CertKeyValue(), defaultConfig.CertKey)
	}

	staging, err := LoadStoredCredentials("staging")
	if err != nil {
		t.Fatalf("LoadStoredCredentials(staging) error = %v", err)
	}
	if staging.EndpointValue() != stagingConfig.PermifyURL {
		t.Fatalf("staging endpoint = %q, want %q", staging.EndpointValue(), stagingConfig.PermifyURL)
	}
	if staging.Token != stagingConfig.Token {
		t.Fatalf("staging token = %q, want %q", staging.Token, stagingConfig.Token)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(credentials) error = %v", err)
	}
	contents := string(data)
	for _, want := range []string{"endpoint:", "token:", "cert_path:", "cert_key:"} {
		if !strings.Contains(contents, want) {
			t.Fatalf("credentials file missing %q:\n%s", want, contents)
		}
	}
	if strings.Contains(contents, "permify_url:") || strings.Contains(contents, "cert_key_path:") {
		t.Fatalf("credentials file used legacy field names:\n%s", contents)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat(credentials) error = %v", err)
		}
		if mode := info.Mode().Perm(); mode != credentialsFileMode {
			t.Fatalf("credentials file mode = %04o, want %04o", mode, credentialsFileMode)
		}
	}
}

func TestLoadStoredCredentialsSupportsLegacyNames(t *testing.T) {
	path := withCredentialsPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	contents := []byte("default:\n  permify_url: https://legacy.example:3478\n  token: legacy-token\n  cert_path: /tmp/legacy.crt\n  cert_key_path: /tmp/legacy.key\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("WriteFile(credentials) error = %v", err)
	}

	stored, err := LoadStoredCredentials("default")
	if err != nil {
		t.Fatalf("LoadStoredCredentials() error = %v", err)
	}
	if stored.EndpointValue() != "https://legacy.example:3478" {
		t.Fatalf("legacy endpoint = %q", stored.EndpointValue())
	}
	if stored.CertKeyValue() != "/tmp/legacy.key" {
		t.Fatalf("legacy cert key = %q", stored.CertKeyValue())
	}
}

func TestLoadAppliesStoredCredentialsAndWriteKeepsSecretsOutOfConfig(t *testing.T) {
	withCredentialsPath(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".permctl")
	if err := os.WriteFile(configPath, []byte("default:\n  permify_url: localhost:3478\n  tenant: t1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}
	if err := WriteStoredCredentials("default", CoreConfig{
		PermifyURL: "https://secure.example:3478",
		Token:      "stored-token",
		CertPath:  "/tmp/client.crt",
		CertKey:   "/tmp/client.key",
	}); err != nil {
		t.Fatalf("WriteStoredCredentials() error = %v", err)
	}

	if err := IsConfigured(configPath, "default"); err != nil {
		t.Fatalf("IsConfigured() error = %v", err)
	}
	if err := Load(configPath, "default"); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if CliConfig.PermifyURL != "https://secure.example:3478" {
		t.Fatalf("CliConfig.PermifyURL = %q", CliConfig.PermifyURL)
	}
	if CliConfig.Tenant != "t1" {
		t.Fatalf("CliConfig.Tenant = %q", CliConfig.Tenant)
	}
	if CliConfig.Token != "stored-token" {
		t.Fatalf("CliConfig.Token = %q", CliConfig.Token)
	}

	if err := Write(); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	contents := string(data)
	for _, forbidden := range []string{"stored-token", "client.crt", "client.key", "https://secure.example"} {
		if strings.Contains(contents, forbidden) {
			t.Fatalf("normal config leaked %q:\n%s", forbidden, contents)
		}
	}
	if !strings.Contains(contents, "tenant: t1") {
		t.Fatalf("normal config lost tenant:\n%s", contents)
	}
}
