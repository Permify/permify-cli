package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func resetConfigState() {
	CliConfig = CoreConfig{}
	profileConfigs = ProfileConfigs{}
}

func createTempCertFiles(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	certFile := filepath.Join(dir, "client.crt")
	keyFile := filepath.Join(dir, "client.key")

	if err := os.WriteFile(certFile, []byte("dummy-cert"), 0600); err != nil {
		t.Fatalf("failed to write temp cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("dummy-key"), 0600); err != nil {
		t.Fatalf("failed to write temp key file: %v", err)
	}
	return certFile, keyFile
}

func TestConfigCredentialStorageRoundTrip(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	certFile, keyFile := createTempCertFiles(t)
	configPath := filepath.Join(t.TempDir(), ".permctl")

	CliConfig = CoreConfig{
		PermifyURL: "localhost:3478",
		Tenant:     "t1",
		Token:      "secret-token-123",
		CertPath:   certFile,
		CertKey:    keyFile,
	}

	if err := New(configPath, "default"); err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	if err := Write(); err != nil {
		t.Fatalf("Write() returned unexpected error: %v", err)
	}

	// Reset in-memory state and reload from disk
	CliConfig = CoreConfig{}
	if err := IsConfigured(configPath, "default"); err != nil {
		t.Fatalf("IsConfigured() returned unexpected error: %v", err)
	}
	if err := Load(configPath, "default"); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if CliConfig.PermifyURL != "localhost:3478" {
		t.Errorf("expected PermifyURL %q, got %q", "localhost:3478", CliConfig.PermifyURL)
	}
	if CliConfig.Tenant != "t1" {
		t.Errorf("expected Tenant %q, got %q", "t1", CliConfig.Tenant)
	}
	if CliConfig.Token != "secret-token-123" {
		t.Errorf("expected Token %q, got %q", "secret-token-123", CliConfig.Token)
	}
	if CliConfig.CertPath != certFile {
		t.Errorf("expected CertPath %q, got %q", certFile, CliConfig.CertPath)
	}
	if CliConfig.CertKey != keyFile {
		t.Errorf("expected CertKey %q, got %q", keyFile, CliConfig.CertKey)
	}
	if !CliConfig.SslEnabled {
		t.Errorf("expected SslEnabled to be true when CertPath and CertKey are configured")
	}

	endpoint, token, gotCertPath, gotCertKey := GetCredentials()
	if endpoint != "localhost:3478" || token != "secret-token-123" || gotCertPath != certFile || gotCertKey != keyFile {
		t.Errorf("GetCredentials() mismatch: got (%q, %q, %q, %q)", endpoint, token, gotCertPath, gotCertKey)
	}
}

func TestConfigFilePermissions0600(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	configPath := filepath.Join(t.TempDir(), ".permctl")
	CliConfig = CoreConfig{
		PermifyURL: "localhost:3478",
		Tenant:     "t1",
		Token:      "initial-token",
	}

	if err := New(configPath, "default"); err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("os.Stat() failed: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("expected file permissions 0600 after New(), got %04o", perm)
		}
	}

	// Simulate an existing file with overly permissive 0644 mode and verify Write() tightens it to 0600
	if err := os.Chmod(configPath, 0644); err != nil {
		t.Fatalf("os.Chmod(0644) failed: %v", err)
	}
	CliConfig.Token = "updated-token"
	if err := Write(); err != nil {
		t.Fatalf("Write() returned unexpected error: %v", err)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("os.Stat() failed: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("expected file permissions 0600 after Write(), got %04o", perm)
		}
	}
}

func TestComputeSslEnabled(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		certPath string
		certKey  string
		want     bool
	}{
		{
			name:     "plaintext endpoint without certs",
			endpoint: "localhost:3478",
			want:     false,
		},
		{
			name:     "http scheme without certs",
			endpoint: "http://localhost:3478",
			want:     false,
		},
		{
			name:     "https scheme without certs",
			endpoint: "https://api.permify.co:3478",
			want:     true,
		},
		{
			name:     "grpcs scheme without certs",
			endpoint: "GRPCS://api.permify.co:3478",
			want:     true,
		},
		{
			name:     "plaintext endpoint with both cert and key",
			endpoint: "localhost:3478",
			certPath: "/path/to/cert.pem",
			certKey:  "/path/to/key.pem",
			want:     true,
		},
		{
			name:     "plaintext endpoint with only cert path",
			endpoint: "localhost:3478",
			certPath: "/path/to/cert.pem",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ComputeSslEnabled(tc.endpoint, tc.certPath, tc.certKey); got != tc.want {
				t.Errorf("ComputeSslEnabled(%q, %q, %q) = %v, want %v", tc.endpoint, tc.certPath, tc.certKey, got, tc.want)
			}
		})
	}
}

func TestConfigValidateErrors(t *testing.T) {
	certFile, keyFile := createTempCertFiles(t)
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		cfg         CoreConfig
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing endpoint",
			cfg:         CoreConfig{PermifyURL: "   "},
			wantErr:     true,
			errContains: "permify endpoint is missing",
		},
		{
			name: "cert_path without cert_key",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				CertPath:   certFile,
			},
			wantErr:     true,
			errContains: "both cert_path and cert_key must be provided",
		},
		{
			name: "cert_key without cert_path",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				CertKey:    keyFile,
			},
			wantErr:     true,
			errContains: "both cert_path and cert_key must be provided",
		},
		{
			name: "non-existent cert_path",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				CertPath:   filepath.Join(tempDir, "missing.crt"),
				CertKey:    keyFile,
			},
			wantErr:     true,
			errContains: "certificate file not found",
		},
		{
			name: "non-existent cert_key",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				CertPath:   certFile,
				CertKey:    filepath.Join(tempDir, "missing.key"),
			},
			wantErr:     true,
			errContains: "certificate key file not found",
		},
		{
			name: "cert_path is a directory",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				CertPath:   tempDir,
				CertKey:    keyFile,
			},
			wantErr:     true,
			errContains: "is a directory",
		},
		{
			name: "cert_key is a directory",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				CertPath:   certFile,
				CertKey:    tempDir,
			},
			wantErr:     true,
			errContains: "is a directory",
		},
		{
			name: "valid config without TLS certs",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				Token:      "my-token",
			},
			wantErr: false,
		},
		{
			name: "valid config with TLS certs",
			cfg: CoreConfig{
				PermifyURL: "localhost:3478",
				Token:      "my-token",
				CertPath:   certFile,
				CertKey:    keyFile,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errContains)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsConfiguredAndLoadInvalidYAML(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	badFile := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(badFile, []byte(":::invalid_yaml\t[\n"), 0600); err != nil {
		t.Fatalf("failed to write malformed config file: %v", err)
	}

	if err := IsConfigured(badFile, "default"); err == nil {
		t.Errorf("expected IsConfigured() to return error on malformed YAML, got nil")
	}
	if err := Load(badFile, "default"); err == nil {
		t.Errorf("expected Load() to return error on malformed YAML, got nil")
	}
}

func TestGetCredentialsEnvOverrides(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	CliConfig = CoreConfig{
		PermifyURL: "stored.permify.local:3478",
		Token:      "stored-token",
		CertPath:   "/stored/cert.pem",
		CertKey:    "/stored/key.pem",
	}

	t.Run("stored credentials without env overrides", func(t *testing.T) {
		ep, tok, cp, ck := GetCredentials()
		if ep != "stored.permify.local:3478" || tok != "stored-token" || cp != "/stored/cert.pem" || ck != "/stored/key.pem" {
			t.Errorf("unexpected stored credentials: (%q, %q, %q, %q)", ep, tok, cp, ck)
		}
	})

	t.Run("primary environment variable overrides", func(t *testing.T) {
		t.Setenv(EnvPermifyEndpoint, "env.permify.local:3478")
		t.Setenv(EnvPermifyToken, "env-token")
		t.Setenv(EnvPermifyCertPath, "/env/cert.pem")
		t.Setenv(EnvPermifyCertKey, "/env/key.pem")

		ep, tok, cp, ck := GetCredentials()
		if ep != "env.permify.local:3478" || tok != "env-token" || cp != "/env/cert.pem" || ck != "/env/key.pem" {
			t.Errorf("unexpected env overridden credentials: (%q, %q, %q, %q)", ep, tok, cp, ck)
		}
	})

	t.Run("alias environment variable overrides", func(t *testing.T) {
		t.Setenv(EnvPermifyURL, "url.permify.local:3478")
		t.Setenv(EnvPermifyCertKeyPath, "/alias/key.pem")

		ep, _, _, ck := GetCredentials()
		if ep != "url.permify.local:3478" || ck != "/alias/key.pem" {
			t.Errorf("unexpected alias overridden credentials: endpoint=%q, certKey=%q", ep, ck)
		}
	})
}
