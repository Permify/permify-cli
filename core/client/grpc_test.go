package client

import (
	"context"
	"path/filepath"
	"testing"
)

func TestTransportCredentialsInsecureWhenNoCert(t *testing.T) {
	creds, err := transportCredentials("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
	if got := creds.Info().SecurityProtocol; got != "insecure" {
		t.Fatalf("expected insecure security protocol, got %q", got)
	}
}

func TestTransportCredentialsErrorsOnMissingCert(t *testing.T) {
	dir := t.TempDir()
	_, err := transportCredentials(filepath.Join(dir, "missing.crt"), filepath.Join(dir, "missing.key"))
	if err == nil {
		t.Fatal("expected error for missing certificate files")
	}
}

func TestTokenAuthMetadata(t *testing.T) {
	auth := tokenAuth{token: "secret"}
	md, err := auth.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md["authorization"] != "Bearer secret" {
		t.Fatalf("unexpected authorization header: %q", md["authorization"])
	}
	if auth.RequireTransportSecurity() {
		t.Fatal("expected RequireTransportSecurity to be false")
	}
}
