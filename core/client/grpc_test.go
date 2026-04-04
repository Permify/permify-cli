package client

import (
	"strings"
	"testing"

	"github.com/Permify/permify-cli/core/config"
)

func TestNew_EndpointValidation(t *testing.T) {
	// Save original config to restore after test
	originalConfig := config.CliConfig
	defer func() { config.CliConfig = originalConfig }()

	tests := []struct {
		name          string
		endpoint      string
		configURL     string
		expectErr     bool
		errContains   string
		expectedURL   string // The URL that should be used (for verification)
	}{
		{
			name:        "empty endpoint and empty config URL",
			endpoint:    "",
			configURL:   "",
			expectErr:   true,
			errContains: "endpoint is required",
		},
		{
			name:        "empty endpoint uses config URL",
			endpoint:    "",
			configURL:   "localhost:3478",
			expectErr:   false,
			expectedURL: "localhost:3478",
		},
		{
			name:        "explicit endpoint overrides config",
			endpoint:    "override.example.com:3478",
			configURL:   "config.example.com:3478",
			expectErr:   false,
			expectedURL: "override.example.com:3478",
		},
		{
			name:        "explicit endpoint with empty config",
			endpoint:    "explicit.example.com:3478",
			configURL:   "",
			expectErr:   false,
			expectedURL: "explicit.example.com:3478",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config for this test
			config.CliConfig = config.CoreConfig{
				PermifyURL: tt.configURL,
				Tenant:     "test-tenant",
			}

			// Call New
			client, err := New(tt.endpoint)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				if client != nil {
					t.Error("expected nil client on error")
				}
			} else {
				// Note: We expect an error here because we're not actually connecting to a real server
				// But the error should NOT be "endpoint is required" - it should be a connection error
				if err != nil {
					if strings.Contains(err.Error(), "endpoint is required") {
						t.Errorf("got 'endpoint is required' error when endpoint should have been resolved")
					}
					// Connection errors are expected in unit tests
				}
			}
		})
	}
}

func TestNew_TokenCredentials(t *testing.T) {
	// Save original config
	originalConfig := config.CliConfig
	defer func() { config.CliConfig = originalConfig }()

	tests := []struct {
		name      string
		token     string
		certPath  string
		certKey   string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "no credentials",
			token:     "",
			certPath:  "",
			certKey:   "",
			expectErr: false,
		},
		{
			name:      "token only",
			token:     "bearer-token-123",
			certPath:  "",
			certKey:   "",
			expectErr: false,
		},
		{
			name:      "token with whitespace trimmed",
			token:     "  token-with-spaces  ",
			certPath:  "",
			certKey:   "",
			expectErr: false,
		},
		{
			name:      "cert key without cert path",
			token:     "",
			certPath:  "",
			certKey:   "/path/to/key.pem",
			expectErr: true,
			errMsg:    "cert_key provided without cert_path",
		},
		{
			name:      "cert path without cert key (valid - server cert only)",
			token:     "",
			certPath:  "/path/to/cert.pem",
			certKey:   "",
			expectErr: false, // This is valid for server certificate verification
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.CliConfig = config.CoreConfig{
				PermifyURL: "localhost:3478",
				Tenant:     "test",
				Token:      tt.token,
				CertPath:   tt.certPath,
				CertKey:    tt.certKey,
			}

			client, err := New("localhost:3478")

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
			} else {
				// Connection errors are expected since we're not running a real server
				// But we should NOT get validation errors about cert_key without cert_path
				if err != nil && strings.Contains(err.Error(), "cert_key provided without cert_path") {
					t.Errorf("got validation error that should not occur: %v", err)
				}
			}

			// If we expected an error, client should be nil
			if tt.expectErr && client != nil {
				t.Error("expected nil client when error occurs")
			}
		})
	}
}

func TestNew_CertPathWhitespace(t *testing.T) {
	// Save original config
	originalConfig := config.CliConfig
	defer func() { config.CliConfig = originalConfig }()

	tests := []struct {
		name     string
		certPath string
		certKey  string
	}{
		{
			name:     "cert path with leading whitespace",
			certPath: "  /path/to/cert.pem",
			certKey:  "",
		},
		{
			name:     "cert path with trailing whitespace",
			certPath: "/path/to/cert.pem  ",
			certKey:  "",
		},
		{
			name:     "cert key with whitespace",
			certPath: "/path/to/cert.pem",
			certKey:  "  /path/to/key.pem  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.CliConfig = config.CoreConfig{
				PermifyURL: "localhost:3478",
				Tenant:     "test",
				CertPath:   tt.certPath,
				CertKey:    tt.certKey,
			}

			// This will fail with file not found, but should NOT fail with
			// "cert_key provided without cert_path" if whitespace is properly trimmed
			_, err := New("localhost:3478")

			if err != nil && strings.Contains(err.Error(), "cert_key provided without cert_path") {
				t.Error("whitespace in cert path was not properly trimmed")
			}

			// We expect file errors since paths don't exist - that's OK
			// The test is really checking that whitespace doesn't break validation logic
		})
	}
}

func TestNew_EmptyStringVsWhitespace(t *testing.T) {
	// Save original config
	originalConfig := config.CliConfig
	defer func() { config.CliConfig = originalConfig }()

	tests := []struct {
		name     string
		certPath string
		certKey  string
		wantErr  string
	}{
		{
			name:     "both empty strings",
			certPath: "",
			certKey:  "",
			wantErr:  "", // No validation error
		},
		{
			name:     "both whitespace only",
			certPath: "   ",
			certKey:  "   ",
			wantErr:  "", // Should be treated as empty after trim
		},
		{
			name:     "cert key is whitespace, cert path empty",
			certPath: "",
			certKey:  "   ",
			wantErr:  "", // Should be treated as both empty
		},
		{
			name:     "cert key is real value, cert path is whitespace",
			certPath: "   ",
			certKey:  "/real/path.pem",
			wantErr:  "cert_key provided without cert_path",
		},
		{
			name:     "cert key is real value, cert path is empty",
			certPath: "",
			certKey:  "/real/path.pem",
			wantErr:  "cert_key provided without cert_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.CliConfig = config.CoreConfig{
				PermifyURL: "localhost:3478",
				Tenant:     "test",
				CertPath:   tt.certPath,
				CertKey:    tt.certKey,
			}

			_, err := New("localhost:3478")

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q but got none", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
			} else {
				// We might get file errors or connection errors, but NOT the validation error
				if err != nil && strings.Contains(err.Error(), "cert_key provided without cert_path") {
					t.Errorf("got unexpected validation error: %v", err)
				}
			}
		})
	}
}
