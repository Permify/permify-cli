package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_AppliesEnvOverrides(t *testing.T) {
	t.Setenv(PermifyURLEnv, "")
	t.Setenv(PermifyTenantEnv, "")
	t.Setenv(PermifyTokenEnv, "")
	t.Setenv(PermifyCertEnv, "")
	t.Setenv(PermifyCertKeyEnv, "")

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "permctl.yaml")
	err := os.WriteFile(cfgPath, []byte(`
default:
  permify_url: "localhost:3478"
  tenant: "t1"
  token: "fromfile"
  cert_path: "/tmp/client.crt"
  cert_key_path: "/tmp/client.key"
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// reset globals between loads
	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}

	if err := Load(cfgPath, "default"); err != nil {
		t.Fatal(err)
	}
	if CliConfig.Token != "fromfile" {
		t.Fatalf("expected token from file, got %q", CliConfig.Token)
	}

	t.Setenv(PermifyTokenEnv, "fromenv")
	t.Setenv(PermifyCertEnv, "/env/client.crt")
	t.Setenv(PermifyCertKeyEnv, "/env/client.key")

	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}
	if err := Load(cfgPath, "default"); err != nil {
		t.Fatal(err)
	}
	if CliConfig.Token != "fromenv" {
		t.Fatalf("expected token from env, got %q", CliConfig.Token)
	}
	if CliConfig.CertPath != "/env/client.crt" {
		t.Fatalf("expected cert_path from env, got %q", CliConfig.CertPath)
	}
	if CliConfig.CertKeyPath != "/env/client.key" {
		t.Fatalf("expected cert_key_path from env, got %q", CliConfig.CertKeyPath)
	}
}

