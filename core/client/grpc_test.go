package client

import (
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

func generateTestCert(t *testing.T, dir string) (string, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Permify Test"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create cert: %v", err)
	}

	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	certOut, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("failed to open cert.pem: %v", err)
	}
	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	_ = certOut.Close()

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
	certPath, keyPath := generateTestCert(t, tmpDir)

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		Token:      "test-token",
		CertPath:   certPath,
		CertKey:    keyPath,
	}

	c, err := New("")
	if err != nil {
		t.Fatalf("New() failed with valid TLS credentials: %v", err)
	}
	if c == nil {
		t.Fatalf("expected non-nil client")
	}
}

func TestNewClientMissingEndpointOrMismatchedCert(t *testing.T) {
	config.CliConfig = config.CoreConfig{}
	if _, err := New(""); err == nil {
		t.Errorf("expected error when both parameter and stored endpoint are empty")
	}

	config.CliConfig = config.CoreConfig{
		PermifyURL: "localhost:3478",
		CertPath:   "/only/cert.pem",
	}
	if _, err := New("localhost:3478"); err == nil {
		t.Errorf("expected error when only CertPath is set without CertKey")
	}
}
