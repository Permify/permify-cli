package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCoreConfig_YAMLSerialization(t *testing.T) {
	tests := []struct {
		name   string
		config CoreConfig
	}{
		{
			name: "all fields populated",
			config: CoreConfig{
				PermifyURL: "https://example.com:3478",
				Tenant:     "test-tenant",
				Token:      "test-token-123",
				CertPath:   "/path/to/cert.pem",
				CertKey:    "/path/to/key.pem",
				SslEnabled: true,
			},
		},
		{
			name: "only required fields",
			config: CoreConfig{
				PermifyURL: "http://localhost:3478",
				Tenant:     "default",
			},
		},
		{
			name: "with token only",
			config: CoreConfig{
				PermifyURL: "https://example.com:3478",
				Tenant:     "prod-tenant",
				Token:      "bearer-token-xyz",
			},
		},
		{
			name: "with cert fields only",
			config: CoreConfig{
				PermifyURL: "https://example.com:3478",
				Tenant:     "secure-tenant",
				CertPath:   "/etc/ssl/cert.pem",
				CertKey:    "/etc/ssl/key.pem",
			},
		},
		{
			name: "empty optional fields",
			config: CoreConfig{
				PermifyURL: "http://test.local:3478",
				Tenant:     "test",
				Token:      "",
				CertPath:   "",
				CertKey:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to YAML
			data, err := yaml.Marshal(tt.config)
			if err != nil {
				t.Fatalf("failed to marshal config: %v", err)
			}

			// Unmarshal back
			var unmarshaled CoreConfig
			err = yaml.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}

			// Verify all persisted fields match
			if unmarshaled.PermifyURL != tt.config.PermifyURL {
				t.Errorf("PermifyURL mismatch: got %q, want %q", unmarshaled.PermifyURL, tt.config.PermifyURL)
			}
			if unmarshaled.Tenant != tt.config.Tenant {
				t.Errorf("Tenant mismatch: got %q, want %q", unmarshaled.Tenant, tt.config.Tenant)
			}
			if unmarshaled.Token != tt.config.Token {
				t.Errorf("Token mismatch: got %q, want %q", unmarshaled.Token, tt.config.Token)
			}
			if unmarshaled.CertPath != tt.config.CertPath {
				t.Errorf("CertPath mismatch: got %q, want %q", unmarshaled.CertPath, tt.config.CertPath)
			}
			if unmarshaled.CertKey != tt.config.CertKey {
				t.Errorf("CertKey mismatch: got %q, want %q", unmarshaled.CertKey, tt.config.CertKey)
			}
			// Note: SslEnabled has yaml:"-" tag, so it won't be persisted
		})
	}
}

func TestConfigPersistence_RoundTrip(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	tests := []struct {
		name   string
		config CoreConfig
	}{
		{
			name: "full config with credentials",
			config: CoreConfig{
				PermifyURL: "https://api.permify.co:3478",
				Tenant:     "production",
				Token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				CertPath:   "/opt/certs/client.crt",
				CertKey:    "/opt/certs/client.key",
			},
		},
		{
			name: "minimal config",
			config: CoreConfig{
				PermifyURL: "localhost:3478",
				Tenant:     "dev",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: create a fresh config file
			profileConfigs := ProfileConfigs{
				Configs: map[string]CoreConfig{
					"default": tt.config,
				},
				File:    configFile,
				Profile: "default",
			}

			data, err := yaml.Marshal(profileConfigs.Configs)
			if err != nil {
				t.Fatalf("failed to marshal config: %v", err)
			}

			err = os.WriteFile(configFile, data, 0644)
			if err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Load and verify
			readData, err := os.ReadFile(configFile)
			if err != nil {
				t.Fatalf("failed to read config file: %v", err)
			}

			var configs map[string]CoreConfig
			err = yaml.Unmarshal(readData, &configs)
			if err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}

			loaded := configs["default"]

			// Verify all fields survived the round-trip
			if loaded.PermifyURL != tt.config.PermifyURL {
				t.Errorf("PermifyURL: got %q, want %q", loaded.PermifyURL, tt.config.PermifyURL)
			}
			if loaded.Tenant != tt.config.Tenant {
				t.Errorf("Tenant: got %q, want %q", loaded.Tenant, tt.config.Tenant)
			}
			if loaded.Token != tt.config.Token {
				t.Errorf("Token: got %q, want %q", loaded.Token, tt.config.Token)
			}
			if loaded.CertPath != tt.config.CertPath {
				t.Errorf("CertPath: got %q, want %q", loaded.CertPath, tt.config.CertPath)
			}
			if loaded.CertKey != tt.config.CertKey {
				t.Errorf("CertKey: got %q, want %q", loaded.CertKey, tt.config.CertKey)
			}

			// Cleanup
			os.Remove(configFile)
		})
	}
}

func TestIsConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	tests := []struct {
		name      string
		configs   map[string]CoreConfig
		profile   string
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config with credentials",
			configs: map[string]CoreConfig{
				"default": {
					PermifyURL: "https://example.com:3478",
					Tenant:     "test-tenant",
					Token:      "test-token",
					CertPath:   "/path/to/cert",
					CertKey:    "/path/to/key",
				},
			},
			profile:   "default",
			expectErr: false,
		},
		{
			name: "valid config without credentials",
			configs: map[string]CoreConfig{
				"default": {
					PermifyURL: "http://localhost:3478",
					Tenant:     "local",
				},
			},
			profile:   "default",
			expectErr: false,
		},
		{
			name: "missing permify url",
			configs: map[string]CoreConfig{
				"default": {
					Tenant:   "test",
					Token:    "token",
					CertPath: "/cert",
				},
			},
			profile:   "default",
			expectErr: true,
			errMsg:    "permify url is empty",
		},
		{
			name: "missing tenant",
			configs: map[string]CoreConfig{
				"default": {
					PermifyURL: "https://example.com:3478",
					Token:      "token",
				},
			},
			profile:   "default",
			expectErr: true,
			errMsg:    "tenant is empty",
		},
		{
			name: "missing profile",
			configs: map[string]CoreConfig{
				"prod": {
					PermifyURL: "https://example.com:3478",
					Tenant:     "prod-tenant",
				},
			},
			profile:   "default",
			expectErr: true,
			errMsg:    "is empty", // Will match either "permify url is empty" or "tenant is empty"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test config
			data, err := yaml.Marshal(tt.configs)
			if err != nil {
				t.Fatalf("failed to marshal config: %v", err)
			}

			err = os.WriteFile(configFile, data, 0644)
			if err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Test IsConfigured
			err = IsConfigured(configFile, tt.profile)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIsConfigured_MissingFile(t *testing.T) {
	nonExistentFile := "/tmp/this-file-does-not-exist-12345.yaml"
	err := IsConfigured(nonExistentFile, "default")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !containsString(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	testConfig := CoreConfig{
		PermifyURL: "https://secure.example.com:3478",
		Tenant:     "production",
		Token:      "prod-token-abc",
		CertPath:   "/etc/ssl/prod.crt",
		CertKey:    "/etc/ssl/prod.key",
	}

	configs := map[string]CoreConfig{
		"prod": testConfig,
	}

	data, err := yaml.Marshal(configs)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config
	err = Load(configFile, "prod")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify global CliConfig was populated
	if CliConfig.PermifyURL != testConfig.PermifyURL {
		t.Errorf("PermifyURL: got %q, want %q", CliConfig.PermifyURL, testConfig.PermifyURL)
	}
	if CliConfig.Tenant != testConfig.Tenant {
		t.Errorf("Tenant: got %q, want %q", CliConfig.Tenant, testConfig.Tenant)
	}
	if CliConfig.Token != testConfig.Token {
		t.Errorf("Token: got %q, want %q", CliConfig.Token, testConfig.Token)
	}
	if CliConfig.CertPath != testConfig.CertPath {
		t.Errorf("CertPath: got %q, want %q", CliConfig.CertPath, testConfig.CertPath)
	}
	if CliConfig.CertKey != testConfig.CertKey {
		t.Errorf("CertKey: got %q, want %q", CliConfig.CertKey, testConfig.CertKey)
	}
	if !CliConfig.SslEnabled {
		t.Error("SslEnabled should be true for https URL")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
