package client

import (
	"strings"
	"testing"
)

func TestBearerTokenAddsBearerPrefix(t *testing.T) {
	got := bearerToken("test-token")
	if got != "Bearer test-token" {
		t.Fatalf("expected bearer prefix, got %q", got)
	}
}

func TestBearerTokenPreservesExistingBearerPrefix(t *testing.T) {
	got := bearerToken("Bearer test-token")
	if got != "Bearer test-token" {
		t.Fatalf("expected existing bearer token to be preserved, got %q", got)
	}
}

func TestNewTransportCredentialsFallsBackToInsecureWithoutCert(t *testing.T) {
	_, secureTransport, err := newTransportCredentials("", "")
	if err != nil {
		t.Fatalf("expected no error without certificate paths, got %v", err)
	}
	if secureTransport {
		t.Fatal("expected insecure transport when no certificate path is configured")
	}
}

func TestNewTransportCredentialsReturnsCertificateLoadErrors(t *testing.T) {
	_, _, err := newTransportCredentials("missing-cert.pem", "")
	if err == nil {
		t.Fatal("expected an error for a missing certificate")
	}
	if !strings.Contains(err.Error(), "missing-cert.pem") {
		t.Fatalf("expected error to include certificate path, got %v", err)
	}
}
