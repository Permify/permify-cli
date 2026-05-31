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

func TestTokenCredentialsAddsBearerPrefix(t *testing.T) {
	creds := tokenCredentials(config.CoreConfig{
		Token:    "secret-token",
		CertPath: "certs/client.pem",
	})

	if creds == nil {
		t.Fatal("tokenCredentials() = nil, want credentials")
	}

	metadata, err := creds.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}

	if got, want := metadata["authorization"], "Bearer secret-token"; got != want {
		t.Fatalf("authorization metadata = %q, want %q", got, want)
	}
	if !creds.RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity() = false, want true for TLS connections")
	}
}

func TestTokenCredentialsKeepsExistingBearerPrefix(t *testing.T) {
	creds := tokenCredentials(config.CoreConfig{
		Token: "Bearer already-prefixed",
	})

	if creds == nil {
		t.Fatal("tokenCredentials() = nil, want credentials")
	}

	metadata, err := creds.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("GetRequestMetadata() error = %v", err)
	}

	if got, want := metadata["authorization"], "Bearer already-prefixed"; got != want {
		t.Fatalf("authorization metadata = %q, want %q", got, want)
	}
	if creds.RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity() = true, want false for insecure connections")
	}
}

func TestTransportCredentialsDefaultsToInsecure(t *testing.T) {
	creds, err := transportCredentials(config.CoreConfig{})
	if err != nil {
		t.Fatalf("transportCredentials() error = %v", err)
	}

	if got, want := creds.Info().SecurityProtocol, "insecure"; got != want {
		t.Fatalf("SecurityProtocol = %q, want %q", got, want)
	}
}

func TestTransportCredentialsUsesTLSCertificate(t *testing.T) {
	certPath, keyPath := writeTestCertificate(t)

	creds, err := transportCredentials(config.CoreConfig{
		CertPath:    certPath,
		CertKeyPath: keyPath,
	})
	if err != nil {
		t.Fatalf("transportCredentials() error = %v", err)
	}

	if got, want := creds.Info().SecurityProtocol, "tls"; got != want {
		t.Fatalf("SecurityProtocol = %q, want %q", got, want)
	}
}

func TestTransportCredentialsRejectsKeyWithoutCertificate(t *testing.T) {
	_, err := transportCredentials(config.CoreConfig{
		CertKeyPath: "certs/client.key",
	})
	if err == nil {
		t.Fatal("transportCredentials() error = nil, want validation error")
	}
}

func TestNormalizeEndpointStripsHTTPAndHTTPSSchemes(t *testing.T) {
	testCases := map[string]string{
		"http://localhost:3478":     "localhost:3478",
		"https://permify.example":   "permify.example",
		"permify.internal:3478":     "permify.internal:3478",
		"https://localhost:3478/v1": "localhost:3478/v1",
	}

	for input, want := range testCases {
		if got := normalizeEndpoint(input); got != want {
			t.Fatalf("normalizeEndpoint(%q) = %q, want %q", input, got, want)
		}
	}
}

func writeTestCertificate(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "client.pem")
	keyPath := filepath.Join(dir, "client.key")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatalf("WriteFile(cert) error = %v", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatalf("WriteFile(key) error = %v", err)
	}

	return certPath, keyPath
}
