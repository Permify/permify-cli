package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestTokenCredentials_Metadata(t *testing.T) {
	secureCreds := secureTokenCredentials{"authorization": "Bearer sec-tok-123"}
	if !secureCreds.RequireTransportSecurity() {
		t.Errorf("secureTokenCredentials should require transport security")
	}
	md, err := secureCreds.GetRequestMetadata(context.Background())
	if err != nil || md["authorization"] != "Bearer sec-tok-123" {
		t.Errorf("unexpected metadata: %v, err: %v", md, err)
	}

	nonSecureCreds := nonSecureTokenCredentials{"authorization": "Bearer non-sec-456"}
	if nonSecureCreds.RequireTransportSecurity() {
		t.Errorf("nonSecureTokenCredentials should not require transport security")
	}
	md2, err := nonSecureCreds.GetRequestMetadata(context.Background())
	if err != nil || md2["authorization"] != "Bearer non-sec-456" {
		t.Errorf("unexpected metadata: %v, err: %v", md2, err)
	}
}

func TestClient_MissingEndpointError(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	os.Unsetenv("PERMIFY_ENDPOINT")

	_, err := New("")
	if err == nil {
		t.Errorf("expected error when endpoint is missing, got nil")
	}
}

func TestClient_CertKeyWithoutCertPath(t *testing.T) {
	config.CliConfig = config.CoreConfig{
		Endpoint: "localhost:3478",
		CertKey:  "/some/key.key",
	}

	_, err := New("")
	if err == nil {
		t.Errorf("expected error when cert_key is provided without cert_path, got nil")
	}
}

func TestClient_NonExistentCertPath(t *testing.T) {
	tmpDir := t.TempDir()
	config.CliConfig = config.CoreConfig{
		Endpoint: "localhost:3478",
		CertPath: filepath.Join(tmpDir, "does_not_exist.crt"),
	}

	_, err := New("")
	if err == nil {
		t.Errorf("expected error for non-existent cert_path, got nil")
	}
}

func TestClient_NewWithConfig(t *testing.T) {
	cfg := config.CoreConfig{
		Endpoint: "127.0.0.1:3478",
		Tenant:   "t1",
		Token:    "test-jwt-bearer",
		Insecure: true,
	}

	client, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatalf("expected client to initialize without error, got: %v", err)
	}
	if client == nil {
		t.Errorf("expected non-nil client")
	}
}
