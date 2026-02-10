package client

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Credentials holds the stored client credentials for connecting to Permify.
type Credentials struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token,omitempty"`
	CertPath string `yaml:"cert_path,omitempty"`
	CertKey  string `yaml:"cert_key,omitempty"`
}

// credentialsDir returns the OS-specific directory for storing credentials.
func credentialsDir() (string, error) {
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		if home == "" {
			return "", errors.New("USERPROFILE environment variable is not set")
		}
		return filepath.Join(home, ".permify"), nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
	}
	return filepath.Join(home, ".permify"), nil
}

// credentialsPath returns the full path to the credentials file.
func credentialsPath() (string, error) {
	dir, err := credentialsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials"), nil
}

// marshalCredentials serializes credentials to YAML bytes.
func marshalCredentials(creds Credentials) ([]byte, error) {
	data, err := yaml.Marshal(creds)
	if err != nil {
		return nil, fmt.Errorf("marshal credentials: %w", err)
	}
	return data, nil
}

// unmarshalCredentials reads and parses credentials from the given file path.
// Returns an empty Credentials struct if the file does not exist.
func unmarshalCredentials(path string) (Credentials, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Credentials{}, nil
	}
	if err != nil {
		return Credentials{}, fmt.Errorf("read credentials file %s: %w", path, err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("parse credentials file %s: %w", path, err)
	}

	return creds, nil
}

// LoadCredentials reads stored credentials from ~/.permify/credentials.
// Returns an empty Credentials struct if the file does not exist.
func LoadCredentials() (Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return Credentials{}, err
	}

	return unmarshalCredentials(path)
}

// StoreCredentials writes the given credentials to ~/.permify/credentials
// with user-only permissions (0600).
func StoreCredentials(creds Credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create credentials directory: %w", err)
	}

	data, err := marshalCredentials(creds)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, fs.FileMode(0600)); err != nil {
		return fmt.Errorf("write credentials file %s: %w", path, err)
	}

	return nil
}
