package credentials

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func tempCredsFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "credentials.yaml")
}

func TestLoadFrom_MissingFile_ReturnsNotFound(t *testing.T) {
	path := tempCredsFile(t)
	_, err := LoadFrom(path, DefaultProfile)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLoadFrom_MissingProfile_ReturnsNotFound(t *testing.T) {
	path := tempCredsFile(t)
	if err := SaveTo(path, "staging", &Credentials{Endpoint: "x"}); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}
	_, err := LoadFrom(path, "production")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing profile, got %v", err)
	}
}

func TestSaveTo_LoadFrom_RoundTrip(t *testing.T) {
	path := tempCredsFile(t)
	want := &Credentials{
		Endpoint: "permify.example.com:443",
		Token:    "super-secret-token",
		CertPath: "/etc/permify/client.crt",
		CertKey:  "/etc/permify/client.key",
	}
	if err := SaveTo(path, DefaultProfile, want); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}
	got, err := LoadFrom(path, DefaultProfile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if *got != *want {
		t.Fatalf("round-trip mismatch:\n want=%+v\n got =%+v", want, got)
	}
}

func TestSaveTo_PreservesOtherProfiles(t *testing.T) {
	path := tempCredsFile(t)
	stg := &Credentials{Endpoint: "stg:443", Token: "stg-tok"}
	prd := &Credentials{Endpoint: "prd:443", Token: "prd-tok"}
	if err := SaveTo(path, "staging", stg); err != nil {
		t.Fatalf("SaveTo staging: %v", err)
	}
	if err := SaveTo(path, "production", prd); err != nil {
		t.Fatalf("SaveTo production: %v", err)
	}
	// Staging should still be present after the second write.
	got, err := LoadFrom(path, "staging")
	if err != nil {
		t.Fatalf("LoadFrom staging: %v", err)
	}
	if *got != *stg {
		t.Fatalf("staging clobbered: want %+v got %+v", stg, got)
	}
	got, err = LoadFrom(path, "production")
	if err != nil {
		t.Fatalf("LoadFrom production: %v", err)
	}
	if *got != *prd {
		t.Fatalf("production mismatch: want %+v got %+v", prd, got)
	}
}

func TestSaveTo_FilePermissionsAre0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission semantics do not apply on Windows")
	}
	path := tempCredsFile(t)
	if err := SaveTo(path, DefaultProfile, &Credentials{Token: "t"}); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("expected file mode 0600, got %#o", mode)
	}
}

func TestLoadFrom_InsecurePermissions_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission semantics do not apply on Windows")
	}
	path := tempCredsFile(t)
	if err := SaveTo(path, DefaultProfile, &Credentials{Token: "t"}); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}
	// Deliberately loosen permissions and check that Load rejects it.
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if _, err := LoadFrom(path, DefaultProfile); err == nil {
		t.Fatalf("expected insecure-permissions error, got nil")
	}
}

func TestSaveTo_NilCredentials_ReturnsError(t *testing.T) {
	path := tempCredsFile(t)
	if err := SaveTo(path, DefaultProfile, nil); err == nil {
		t.Fatalf("expected error saving nil credentials, got nil")
	}
}

func TestDeleteFrom_RemovesProfile(t *testing.T) {
	path := tempCredsFile(t)
	if err := SaveTo(path, "staging", &Credentials{Token: "s"}); err != nil {
		t.Fatalf("SaveTo staging: %v", err)
	}
	if err := SaveTo(path, "production", &Credentials{Token: "p"}); err != nil {
		t.Fatalf("SaveTo production: %v", err)
	}
	if err := DeleteFrom(path, "staging"); err != nil {
		t.Fatalf("DeleteFrom: %v", err)
	}
	if _, err := LoadFrom(path, "staging"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected staging gone, got err=%v", err)
	}
	// Production survives.
	got, err := LoadFrom(path, "production")
	if err != nil {
		t.Fatalf("LoadFrom production after delete: %v", err)
	}
	if got.Token != "p" {
		t.Fatalf("production token clobbered: got %q", got.Token)
	}
}

func TestDeleteFrom_LastProfileRemovesFile(t *testing.T) {
	path := tempCredsFile(t)
	if err := SaveTo(path, DefaultProfile, &Credentials{Token: "t"}); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}
	if err := DeleteFrom(path, DefaultProfile); err != nil {
		t.Fatalf("DeleteFrom: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected credentials file removed, stat err=%v", err)
	}
}

func TestDeleteFrom_MissingFileIsNoOp(t *testing.T) {
	path := tempCredsFile(t)
	if err := DeleteFrom(path, DefaultProfile); err != nil {
		t.Fatalf("expected nil error deleting from missing file, got %v", err)
	}
}

func TestGetCredentialsPath_HonorsEnvOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), "custom.yaml")
	t.Setenv(envFileOverride, want)
	got, err := GetCredentialsPath()
	if err != nil {
		t.Fatalf("GetCredentialsPath: %v", err)
	}
	if got != want {
		t.Fatalf("env override ignored: want %q got %q", want, got)
	}
}

func TestGetCredentialsPath_DefaultsToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv(envFileOverride, "")
	t.Setenv("HOME", home)
	// On Windows, os.UserHomeDir uses USERPROFILE; pin it to home so the
	// fallback also resolves deterministically for this test.
	t.Setenv("USERPROFILE", home)
	got, err := GetCredentialsPath()
	if err != nil {
		t.Fatalf("GetCredentialsPath: %v", err)
	}
	want := filepath.Join(home, defaultFileName)
	if got != want {
		t.Fatalf("default path mismatch: want %q got %q", want, got)
	}
}

func TestCredentials_IsZero(t *testing.T) {
	if !(Credentials{}).IsZero() {
		t.Fatalf("empty Credentials should be zero")
	}
	if (Credentials{Endpoint: "x"}).IsZero() {
		t.Fatalf("non-empty Credentials should not be zero")
	}
}

func TestLoadFrom_DefaultProfileWhenEmpty(t *testing.T) {
	path := tempCredsFile(t)
	if err := SaveTo(path, "", &Credentials{Endpoint: "e"}); err != nil {
		t.Fatalf("SaveTo with empty profile: %v", err)
	}
	got, err := LoadFrom(path, "")
	if err != nil {
		t.Fatalf("LoadFrom with empty profile: %v", err)
	}
	if got.Endpoint != "e" {
		t.Fatalf("default-profile round-trip mismatch: got %+v", got)
	}
}
