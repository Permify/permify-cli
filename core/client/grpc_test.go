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
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		wantTarget string
		wantTLS    bool
	}{
		{name: "plain host", endpoint: "localhost:3478", wantTarget: "localhost:3478"},
		{name: "https scheme", endpoint: "https://api.example.com:443", wantTarget: "api.example.com:443", wantTLS: true},
		{name: "grpcs scheme", endpoint: "grpcs://api.example.com:3478", wantTarget: "api.example.com:3478", wantTLS: true},
		{name: "http scheme", endpoint: "http://127.0.0.1:3478", wantTarget: "127.0.0.1:3478"},
		{name: "grpc scheme", endpoint: "grpc://127.0.0.1:3478", wantTarget: "127.0.0.1:3478"},
		{name: "trim whitespace", endpoint: "  localhost:3478  ", wantTarget: "localhost:3478"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseEndpoint(test.endpoint)
			if got.Target != test.wantTarget || got.UseTLS != test.wantTLS {
				t.Fatalf("parseEndpoint(%q) = %#v, want target=%q tls=%t", test.endpoint, got, test.wantTarget, test.wantTLS)
			}
		})
	}
}

func TestTrimBearerPrefix(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{token: "", want: ""},
		{token: "token", want: "token"},
		{token: "Bearer abc", want: "abc"},
		{token: "bearer   abc ", want: "abc"},
	}

	for _, test := range tests {
		if got := trimBearerPrefix(test.token); got != test.want {
			t.Fatalf("trimBearerPrefix(%q) = %q, want %q", test.token, got, test.want)
		}
	}
}

func TestTokenCredentialsMetadata(t *testing.T) {
	secure := secureTokenCredentials{"authorization": "Bearer secure-token"}
	if !secure.RequireTransportSecurity() {
		t.Fatalf("secureTokenCredentials should require transport security")
	}
	metadata, err := secure.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("secure metadata error = %v", err)
	}
	if metadata["authorization"] != "Bearer secure-token" {
		t.Fatalf("secure authorization metadata = %q", metadata["authorization"])
	}

	nonSecure := nonSecureTokenCredentials{"authorization": "Bearer local-token"}
	if nonSecure.RequireTransportSecurity() {
		t.Fatalf("nonSecureTokenCredentials should allow insecure local transport")
	}
	metadata, err = nonSecure.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("non-secure metadata error = %v", err)
	}
	if metadata["authorization"] != "Bearer local-token" {
		t.Fatalf("non-secure authorization metadata = %q", metadata["authorization"])
	}
}

func TestDialOptionsRequireBothCertFiles(t *testing.T) {
	_, err := dialOptions(Params{
		Endpoint: "localhost:3478",
		CertPath: "/tmp/client.crt",
	})
	if err == nil || !strings.Contains(err.Error(), "both cert_path and cert_key") {
		t.Fatalf("dialOptions() error = %v, want missing cert key error", err)
	}

	_, err = dialOptions(Params{
		Endpoint: "localhost:3478",
		CertKey:  "/tmp/client.key",
	})
	if err == nil || !strings.Contains(err.Error(), "both cert_path and cert_key") {
		t.Fatalf("dialOptions() error = %v, want missing cert path error", err)
	}
}

func TestDialOptionsAllowBearerTokenOnTLSAndInsecureEndpoints(t *testing.T) {
	options, err := dialOptions(Params{
		Endpoint: "https://permify.example:3478",
		Token:    "Bearer secret",
	})
	if err != nil {
		t.Fatalf("dialOptions(tls token) error = %v", err)
	}
	if len(options) < 2 {
		t.Fatalf("dialOptions(tls token) returned %d options, want at least 2", len(options))
	}

	options, err = dialOptions(Params{
		Endpoint: "localhost:3478",
		Token:    "secret",
	})
	if err != nil {
		t.Fatalf("dialOptions(insecure token) error = %v", err)
	}
	if len(options) < 2 {
		t.Fatalf("dialOptions(insecure token) returned %d options, want at least 2", len(options))
	}
}

func TestDialOptionsAllowMutualTLSCertPair(t *testing.T) {
	certPath, keyPath := writeTestCertificate(t)

	options, err := dialOptions(Params{
		Endpoint: "https://permify.example:3478",
		Token:    "secret",
		CertPath: certPath,
		CertKey:  keyPath,
	})
	if err != nil {
		t.Fatalf("dialOptions(mtls) error = %v", err)
	}
	if len(options) < 2 {
		t.Fatalf("dialOptions(mtls) returned %d options, want at least 2", len(options))
	}
}

func TestNewRejectsEmptyEndpoint(t *testing.T) {
	_, err := New(Params{})
	if err == nil || !strings.Contains(err.Error(), "endpoint is empty") {
		t.Fatalf("New(empty) error = %v, want endpoint error", err)
	}
}

func writeTestCertificate(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "permify-cli-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "client.crt")
	keyPath := filepath.Join(dir, "client.key")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(cert) error = %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("WriteFile(key) error = %v", err)
	}
	return certPath, keyPath
}
