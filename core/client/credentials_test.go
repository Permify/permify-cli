package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreAndLoadCredentials(t *testing.T) {
	// Use a temporary directory to avoid affecting real credentials
	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, ".permify", "credentials")

	// Override credentialsPath for testing by storing directly
	dir := filepath.Dir(credFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	input := Credentials{
		Endpoint: "localhost:3478",
		Token:    "test-token-123",
		CertPath: "/path/to/cert.pem",
		CertKey:  "/path/to/key.pem",
	}

	// Marshal and write manually to the temp path
	data, err := marshalCredentials(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	if err := os.WriteFile(credFile, data, 0600); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Read back and verify
	loaded, err := unmarshalCredentials(credFile)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Endpoint != input.Endpoint {
		t.Errorf("endpoint mismatch: got %q, want %q", loaded.Endpoint, input.Endpoint)
	}
	if loaded.Token != input.Token {
		t.Errorf("token mismatch: got %q, want %q", loaded.Token, input.Token)
	}
	if loaded.CertPath != input.CertPath {
		t.Errorf("cert_path mismatch: got %q, want %q", loaded.CertPath, input.CertPath)
	}
	if loaded.CertKey != input.CertKey {
		t.Errorf("cert_key mismatch: got %q, want %q", loaded.CertKey, input.CertKey)
	}
}

func TestLoadCredentialsMissingFile(t *testing.T) {
	// unmarshalCredentials from non-existent file should return empty struct
	creds, err := unmarshalCredentials("/nonexistent/path/credentials")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if creds.Endpoint != "" || creds.Token != "" {
		t.Errorf("expected empty credentials, got: %+v", creds)
	}
}

func TestLoadCredentialsEmptyValues(t *testing.T) {
	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "credentials")

	input := Credentials{
		Endpoint: "localhost:3478",
	}

	data, err := marshalCredentials(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	if err := os.WriteFile(credFile, data, 0600); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	loaded, err := unmarshalCredentials(credFile)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Endpoint != "localhost:3478" {
		t.Errorf("endpoint mismatch: got %q", loaded.Endpoint)
	}
	if loaded.Token != "" {
		t.Errorf("expected empty token, got %q", loaded.Token)
	}
	if loaded.CertPath != "" {
		t.Errorf("expected empty cert_path, got %q", loaded.CertPath)
	}
	if loaded.CertKey != "" {
		t.Errorf("expected empty cert_key, got %q", loaded.CertKey)
	}
}

func TestIsSecureEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
		want     bool
	}{
		{"https://api.permify.co", true},
		{"HTTPS://api.permify.co", true},
		{"grpcs://api.permify.co", true},
		{"http://localhost:3478", false},
		{"localhost:3478", false},
		{"grpc://localhost:3478", false},
	}

	for _, tt := range tests {
		got := isSecureEndpoint(tt.endpoint)
		if got != tt.want {
			t.Errorf("isSecureEndpoint(%q) = %v, want %v", tt.endpoint, got, tt.want)
		}
	}
}
