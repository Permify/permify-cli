package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCoreConfig_YAMLSerialization(t *testing.T) {
	cfg := CoreConfig{
		Endpoint: "permify.corp.internal:443",
		Tenant:   "t_test_123",
		Token:    "bearer-token-secret-xyz",
		CertPath: "/etc/ssl/certs/client.crt",
		CertKey:  "/etc/ssl/private/client.key",
		Insecure: false,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	var unmarshaled CoreConfig
	err = yaml.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if unmarshaled.Endpoint != cfg.Endpoint {
		t.Errorf("expected endpoint %s, got %s", cfg.Endpoint, unmarshaled.Endpoint)
	}
	if unmarshaled.Token != cfg.Token {
		t.Errorf("expected token %s, got %s", cfg.Token, unmarshaled.Token)
	}
	if unmarshaled.CertPath != cfg.CertPath {
		t.Errorf("expected cert_path %s, got %s", cfg.CertPath, unmarshaled.CertPath)
	}
	if unmarshaled.CertKey != cfg.CertKey {
		t.Errorf("expected cert_key %s, got %s", cfg.CertKey, unmarshaled.CertKey)
	}
}

func TestCoreConfig_GetEndpoint(t *testing.T) {
	c1 := CoreConfig{Endpoint: "endpoint.local:3478"}
	if c1.GetEndpoint() != "endpoint.local:3478" {
		t.Errorf("expected endpoint.local:3478, got %s", c1.GetEndpoint())
	}

	c2 := CoreConfig{PermifyURL: "permify.local:3478"}
	if c2.GetEndpoint() != "permify.local:3478" {
		t.Errorf("expected permify.local:3478, got %s", c2.GetEndpoint())
	}

	c3 := CoreConfig{}
	c3.SetEndpoint("updated.local:3478")
	if c3.Endpoint != "updated.local:3478" || c3.PermifyURL != "updated.local:3478" {
		t.Errorf("SetEndpoint failed to update both fields")
	}
}

func TestCoreConfig_ValidateCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "test.crt")
	keyFile := filepath.Join(tmpDir, "test.key")

	err := os.WriteFile(certFile, []byte("CERT_DATA"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(keyFile, []byte("KEY_DATA"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	validCfg := CoreConfig{
		CertPath: certFile,
		CertKey:  keyFile,
	}
	if err := validCfg.ValidateCredentials(); err != nil {
		t.Errorf("expected valid credentials, got error: %v", err)
	}

	invalidCfg := CoreConfig{
		CertPath: filepath.Join(tmpDir, "non_existent.crt"),
	}
	if err := invalidCfg.ValidateCredentials(); err == nil {
		t.Errorf("expected error for non-existent cert_path, got nil")
	}
}

func TestCoreConfig_LoadAndPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "permctl.yaml")

	CliConfig = CoreConfig{
		Endpoint: "127.0.0.1:3478",
		Tenant:   "t1",
		Token:    "test-token",
	}

	err := New(configFile, "default")
	if err != nil {
		t.Fatalf("failed to create new config: %v", err)
	}

	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	// Verify that file was created
	if info.Size() == 0 {
		t.Errorf("config file is empty")
	}

	err = IsConfigured(configFile, "default")
	if err != nil {
		t.Errorf("expected IsConfigured to succeed, got %v", err)
	}

	err = Load(configFile, "default")
	if err != nil {
		t.Errorf("expected Load to succeed, got %v", err)
	}

	if CliConfig.Token != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", CliConfig.Token)
	}
}
