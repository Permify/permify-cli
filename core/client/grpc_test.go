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

func TestTokenCredentials(t *testing.T) {
	secureCreds := tokenCredentials("secret", true)
	if !secureCreds.RequireTransportSecurity() {
		t.Fatalf("expected secure credentials to require transport security")
	}
	secureMetadata, err := secureCreds.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}
	if got := secureMetadata["authorization"]; got != "Bearer secret" {
		t.Fatalf("unexpected secure metadata authorization: got %q", got)
	}

	insecureCreds := tokenCredentials("secret", false)
	if insecureCreds.RequireTransportSecurity() {
		t.Fatalf("expected insecure credentials to not require transport security")
	}
	insecureMetadata, err := insecureCreds.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}
	if got := insecureMetadata["authorization"]; got != "Bearer secret" {
		t.Fatalf("unexpected insecure metadata authorization: got %q", got)
	}
}

func TestTransportCredentials(t *testing.T) {
	originalCredentials := config.CliCredentials
	originalConfig := config.CliConfig
	t.Cleanup(func() {
		config.CliCredentials = originalCredentials
		config.CliConfig = originalConfig
	})

	t.Run("insecure", func(t *testing.T) {
		config.CliConfig.SslEnabled = false
		creds, err := transportCredentials(false)
		if err != nil {
			t.Fatalf("transportCredentials(false) error = %v", err)
		}
		if creds == nil {
			t.Fatalf("expected credentials, got nil")
		}
	})

	t.Run("tls pair", func(t *testing.T) {
		certPath, certKey := generateTestCertificatePair(t)
		config.CliConfig.SslEnabled = true
		config.CliCredentials.CertPath = certPath
		config.CliCredentials.CertKey = certKey

		creds, err := transportCredentials(true)
		if err != nil {
			t.Fatalf("transportCredentials(true) error = %v", err)
		}
		if creds == nil {
			t.Fatalf("expected credentials, got nil")
		}
	})

	t.Run("missing cert key", func(t *testing.T) {
		config.CliConfig.SslEnabled = true
		config.CliCredentials.CertPath = "/tmp/client.pem"
		config.CliCredentials.CertKey = ""

		_, err := transportCredentials(true)
		if err == nil {
			t.Fatalf("expected error for missing cert key")
		}
	})
}

func generateTestCertificatePair(t *testing.T) (string, string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "permify-cli-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "client.crt")
	certKey := filepath.Join(tempDir, "client.key")

	certFile, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("os.Create(cert) error = %v", err)
	}
	defer certFile.Close()
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		t.Fatalf("pem.Encode(cert) error = %v", err)
	}

	keyFile, err := os.Create(certKey)
	if err != nil {
		t.Fatalf("os.Create(key) error = %v", err)
	}
	defer keyFile.Close()
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		t.Fatalf("pem.Encode(key) error = %v", err)
	}

	return certPath, certKey
}
