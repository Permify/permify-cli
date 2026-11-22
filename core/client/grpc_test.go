package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Permify/permify-cli/core/config"
)

func generateTestCertPair(t *testing.T, dir string) (string, string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Permify CLI Test"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create X.509 certificate: %v", err)
	}

	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	return certPath, keyPath
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		want      string
		wantTLS   bool
		wantErr   bool
	}{
		{
			name:     "bare host and port",
			endpoint: "localhost:3478",
			want:     "localhost:3478",
			wantTLS:  false,
		},
		{
			name:     "http scheme",
			endpoint: "http://localhost:3478",
			want:     "localhost:3478",
			wantTLS:  false,
		},
		{
			name:     "http scheme with trailing slash",
			endpoint: "http://localhost:3478/",
			want:     "localhost:3478",
			wantTLS:  false,
		},
		{
			name:     "https scheme",
			endpoint: "https://permify.example.com:3478",
			want:     "permify.example.com:3478",
			wantTLS:  true,
		},
		{
			name:     "uppercase HTTPS scheme with trailing slash",
			endpoint: "HTTPS://permify.example.com:3478/",
			want:     "permify.example.com:3478",
			wantTLS:  true,
		},
		{
			name:     "grpc scheme",
			endpoint: "grpc://localhost:3478",
			want:     "localhost:3478",
			wantTLS:  false,
		},
		{
			name:     "grpcs scheme",
			endpoint: "grpcs://permify.example.com:3478/",
			want:     "permify.example.com:3478",
			wantTLS:  true,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:     "whitespace only endpoint",
			endpoint: "   ",
			wantErr:  true,
		},
		{
			name:     "scheme only endpoint",
			endpoint: "https://",
			wantErr:  true,
		},
		{
			name:     "scheme and slash only endpoint",
			endpoint: "http:///",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotTLS, err := NormalizeEndpoint(tc.endpoint)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeEndpoint(%q) expected error, got nil", tc.endpoint)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeEndpoint(%q) unexpected error: %v", tc.endpoint, err)
			}
			if got != tc.want || gotTLS != tc.wantTLS {
				t.Errorf("NormalizeEndpoint(%q) = (%q, %v), want (%q, %v)", tc.endpoint, got, gotTLS, tc.want, tc.wantTLS)
			}
		})
	}
}

func TestResolveDialOptions_MissingEndpoint(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	t.Cleanup(func() {
		config.CliConfig = config.CoreConfig{}
	})

	if _, _, err := ResolveDialOptions("", "", "", ""); err == nil {
		t.Errorf("expected ResolveDialOptions to return error when endpoint is empty, got nil")
	}

	if _, err := New(""); err == nil {
		t.Errorf("expected New(\"\") to return error when no endpoint is configured, got nil")
	}
}

func TestResolveDialOptions_InsecureWithToken(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	t.Cleanup(func() {
		config.CliConfig = config.CoreConfig{}
	})

	target, opts, err := ResolveDialOptions("http://localhost:3478", "my-secret-token", "", "")
	if err != nil {
		t.Fatalf("ResolveDialOptions returned unexpected error: %v", err)
	}
	if target != "localhost:3478" {
		t.Errorf("expected normalized target %q, got %q", "localhost:3478", target)
	}
	if len(opts) != 2 {
		t.Errorf("expected 2 dial options (transport + token), got %d", len(opts))
	}

	if got := formatBearerToken("my-secret-token"); got != "Bearer my-secret-token" {
		t.Errorf("formatBearerToken mismatch: got %q", got)
	}
	if got := formatBearerToken("Bearer my-secret-token"); got != "Bearer my-secret-token" {
		t.Errorf("formatBearerToken with existing prefix mismatch: got %q", got)
	}
	if got := formatBearerToken("bearer   my-secret-token "); got != "Bearer my-secret-token" {
		t.Errorf("formatBearerToken with lowercase prefix mismatch: got %q", got)
	}

	nonSec := nonSecureTokenCredentials{"authorization": formatBearerToken("my-secret-token")}
	if nonSec.RequireTransportSecurity() {
		t.Errorf("expected nonSecureTokenCredentials.RequireTransportSecurity() to be false")
	}
	md, err := nonSec.GetRequestMetadata(context.Background())
	if err != nil || md["authorization"] != "Bearer my-secret-token" {
		t.Errorf("unexpected metadata from nonSecureTokenCredentials: %v, err=%v", md, err)
	}

	// Verify client.New succeeds with plaintext + nonSecureTokenCredentials
	config.CliConfig = config.CoreConfig{
		PermifyURL: "http://localhost:3478",
		Token:      "my-secret-token",
	}
	c, err := New("")
	if err != nil {
		t.Fatalf("New(\"\") with plaintext endpoint and token failed: %v", err)
	}
	if c == nil {
		t.Fatalf("expected non-nil permify.Client")
	}
}

func TestResolveDialOptions_MismatchedOrMissingCertFiles(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	t.Cleanup(func() {
		config.CliConfig = config.CoreConfig{}
	})

	dir1 := t.TempDir()
	dir2 := t.TempDir()
	cert1, key1 := generateTestCertPair(t, dir1)
	_, key2 := generateTestCertPair(t, dir2)

	invalidCert := filepath.Join(dir1, "invalid.crt")
	invalidKey := filepath.Join(dir1, "invalid.key")
	if err := os.WriteFile(invalidCert, []byte("not-a-valid-pem"), 0600); err != nil {
		t.Fatalf("failed to write invalid cert: %v", err)
	}
	if err := os.WriteFile(invalidKey, []byte("not-a-valid-key"), 0600); err != nil {
		t.Fatalf("failed to write invalid key: %v", err)
	}

	tests := []struct {
		name        string
		certPath    string
		certKey     string
		errContains string
	}{
		{
			name:        "cert_path without cert_key",
			certPath:    cert1,
			certKey:     "",
			errContains: "both cert_path and cert_key must be provided",
		},
		{
			name:        "cert_key without cert_path",
			certPath:    "",
			certKey:     key1,
			errContains: "both cert_path and cert_key must be provided",
		},
		{
			name:        "missing cert_path file",
			certPath:    filepath.Join(dir1, "nonexistent.crt"),
			certKey:     key1,
			errContains: "certificate file not found",
		},
		{
			name:        "missing cert_key file",
			certPath:    cert1,
			certKey:     filepath.Join(dir1, "nonexistent.key"),
			errContains: "certificate key file not found",
		},
		{
			name:        "cert_path is a directory",
			certPath:    dir1,
			certKey:     key1,
			errContains: "is a directory",
		},
		{
			name:        "cert_key is a directory",
			certPath:    cert1,
			certKey:     dir1,
			errContains: "is a directory",
		},
		{
			name:        "invalid PEM data",
			certPath:    invalidCert,
			certKey:     invalidKey,
			errContains: "failed to load TLS certificate key pair",
		},
		{
			name:        "mismatched certificate and private key",
			certPath:    cert1,
			certKey:     key2,
			errContains: "failed to load TLS certificate key pair",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ResolveDialOptions("localhost:3478", "token", tc.certPath, tc.certKey)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errContains)
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
			}
		})
	}
}

func TestResolveDialOptions_ValidTLSAndToken(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	t.Cleanup(func() {
		config.CliConfig = config.CoreConfig{}
	})

	certPath, keyPath := generateTestCertPair(t, t.TempDir())

	target, opts, err := ResolveDialOptions("https://localhost:3478", "tls-token", certPath, keyPath)
	if err != nil {
		t.Fatalf("ResolveDialOptions with valid TLS cert pair returned error: %v", err)
	}
	if target != "localhost:3478" {
		t.Errorf("expected target %q, got %q", "localhost:3478", target)
	}
	if len(opts) != 2 {
		t.Errorf("expected 2 dial options (TLS + secureTokenCredentials), got %d", len(opts))
	}

	sec := secureTokenCredentials{"authorization": formatBearerToken("tls-token")}
	if !sec.RequireTransportSecurity() {
		t.Errorf("expected secureTokenCredentials.RequireTransportSecurity() to be true")
	}
	md, err := sec.GetRequestMetadata(context.Background())
	if err != nil || md["authorization"] != "Bearer tls-token" {
		t.Errorf("unexpected metadata from secureTokenCredentials: %v, err=%v", md, err)
	}

	// Verify client.New succeeds with stored TLS credentials and token in config.CliConfig
	config.CliConfig = config.CoreConfig{
		PermifyURL: "grpcs://localhost:3478",
		Token:      "tls-token",
		CertPath:   certPath,
		CertKey:    keyPath,
		SslEnabled: true,
	}

	c, err := New("")
	if err != nil {
		t.Fatalf("New(\"\") with stored TLS credentials and token failed: %v", err)
	}
	if c == nil {
		t.Fatalf("expected non-nil permify.Client")
	}

	// Verify HTTPS scheme without client certificate files also enables TLS and works with secureTokenCredentials
	config.CliConfig = config.CoreConfig{
		PermifyURL: "https://localhost:3478",
		Token:      "https-only-token",
		SslEnabled: true,
	}
	cHTTPS, err := New("")
	if err != nil {
		t.Fatalf("New(\"\") with HTTPS scheme and token failed: %v", err)
	}
	if cHTTPS == nil {
		t.Fatalf("expected non-nil permify.Client for HTTPS endpoint")
	}
}
