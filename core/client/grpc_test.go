package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"  localhost:3478  ", "localhost:3478"},
		{"https://example.com:3478", "example.com:3478"},
		{"https://example.com:3478/some/path?x=1", "example.com:3478"},
		{"http://127.0.0.1:3478", "127.0.0.1:3478"},
	}
	for _, tt := range tests {
		if got := normalizeEndpoint(tt.in); got != tt.want {
			t.Fatalf("normalizeEndpoint(%q)=%q want=%q", tt.in, got, tt.want)
		}
	}
}

func TestNewFromConfig_RequiresCertAndKeyTogether(t *testing.T) {
	_, err := NewFromConfig(config.CoreConfig{
		PermifyURL:  "localhost:3478",
		CertPath:    "x.crt",
		CertKeyPath: "",
	})
	if err == nil {
		t.Fatalf("expected error when only cert_path is set")
	}
}

func TestTransportCredentials_TLSWithoutFilesFromSslFlag(t *testing.T) {
	creds, err := transportCredentials(config.CoreConfig{
		PermifyURL:  "https://example.com:3478",
		SslEnabled:  true,
		CertPath:    "",
		CertKeyPath: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Info().SecurityProtocol == "insecure" {
		t.Fatalf("expected TLS creds, got insecure")
	}
}

func TestTransportCredentials_TLSWithoutFilesFromEndpointScheme(t *testing.T) {
	creds, err := transportCredentials(config.CoreConfig{
		PermifyURL: "https://example.com:3478",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Info().SecurityProtocol == "insecure" {
		t.Fatalf("expected TLS creds, got insecure")
	}
}

func TestTransportCredentials_BadCertFiles(t *testing.T) {
	tmp := t.TempDir()
	certPath := filepath.Join(tmp, "client.crt")
	keyPath := filepath.Join(tmp, "client.key")
	if err := os.WriteFile(certPath, []byte("not a cert"), 0600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("not a key"), 0600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	_, err := transportCredentials(config.CoreConfig{
		SslEnabled:  true,
		CertPath:    certPath,
		CertKeyPath: keyPath,
	})
	if err == nil {
		t.Fatalf("expected error for invalid cert/key pair")
	}
}
