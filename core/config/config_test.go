package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreConfigStoresCredentialFields(t *testing.T) {
	cfg := CoreConfig{
		PermifyURL: "https://permify.example.com",
		Tenant:     "tenant-1",
		Token:      "secret-token",
		CertPath:   "/tmp/client.crt",
		CertKey:    "/tmp/client.key",
	}

	profileConfigs = ProfileConfigs{
		Configs: map[string]CoreConfig{"default": cfg},
		Profile: "default",
	}

	CliConfig = profileConfigs.Configs["default"]

	if CliConfig.Token != "secret-token" {
		t.Fatalf("expected token to be stored on CoreConfig")
	}
	if CliConfig.CertPath != "/tmp/client.crt" {
		t.Fatalf("expected cert path to be stored on CoreConfig")
	}
	if CliConfig.CertKey != "/tmp/client.key" {
		t.Fatalf("expected cert key to be stored on CoreConfig")
	}
}

func TestConfigFilesAreWrittenOwnerOnlyBecauseTheyMayContainTokens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	CliConfig = CoreConfig{
		PermifyURL: "https://permify.example.com",
		Tenant:     "tenant-1",
		Token:      "secret-token",
		CertPath:   "/tmp/client.crt",
		CertKey:    "/tmp/client.key",
	}

	if err := New(path, "default"); err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	assertOwnerOnly(t, path)

	CliConfig.Token = "new-token"
	if err := Write(); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	assertOwnerOnly(t, path)

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	for _, want := range []string{"token: new-token", "cert_path: /tmp/client.crt", "cert_key: /tmp/client.key"} {
		if !strings.Contains(string(contents), want) {
			t.Fatalf("expected config file to contain %q, got:\n%s", want, contents)
		}
	}
}

func TestWriteTightensExistingConfigPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("default:\n  permify_url: http://localhost\n  tenant: t1\n"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	profileConfigs = ProfileConfigs{
		Configs: map[string]CoreConfig{"default": {PermifyURL: "http://localhost", Tenant: "t1", Token: "secret-token"}},
		File:    path,
		Profile: "default",
	}
	CliConfig = profileConfigs.Configs["default"]

	if err := Write(); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	assertOwnerOnly(t, path)
}

func assertOwnerOnly(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("expected config file permissions 0600, got %o", got)
	}
}
