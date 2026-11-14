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
	"testing"
	"time"

	"github.com/Permify/permify-cli/core/config"
)

func generateTestCertAndKey(t *testing.T, dir string) (string, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Permify Test"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create cert: %v", err)
	}

	certPath := filepath.Join(dir, "cert.pem")
	certOut, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("failed to open cert.pem: %v", err)
	}
	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	_ = certOut.Close()

	keyPath := filepath.Join(dir, "key.pem")
	keyOut, err := os.Create(keyPath)
	if err != nil {
		t.Fatalf("failed to open key.pem: %v", err)
	}
	_ = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	_ = keyOut.Close()

	return certPath, keyPath
}

func TestNewClientWithStoredCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	certPath, keyPath := generateTestCertAndKey(t, tmpDir)

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		Tenant:     "t1",
		Token:      "test-bearer-token",
		CertPath:   certPath,
		CertKey:    keyPath,
	}

	c, err := New("")
	if err != nil {
		t.Fatalf("expected New() to succeed with stored credentials, got error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientMissingOrInvalidCredentials(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	if _, err := New(""); err == nil {
		t.Fatal("expected error when endpoint is missing")
	}

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		CertKey:    "some-key.pem",
	}
	if _, err := New("localhost:3478"); err == nil {
		t.Fatal("expected error when cert_key is set without cert_path")
	}

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		CertPath:   "/nonexistent/cert.pem",
	}
	if _, err := New("localhost:3478"); err == nil {
		t.Fatal("expected error when cert_path does not exist")
	}
}

func TestTokenCredentialsMetadata(t *testing.T) {
	tc := tokenCredentials{token: "abc-123"}
	md, err := tc.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md["authorization"] != "Bearer abc-123" {
		t.Errorf("expected Bearer abc-123, got %s", md["authorization"])
	}
}
