package client

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Credentials represents persisted client configuration for the Permify CLI.
type Credentials struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
	CertPath string `yaml:"cert_path"`
	CertKey  string `yaml:"cert_key"`
}

// credentialsFilePath resolves the OS-specific location for the credentials file.
func credentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, ".permify", "credentials"), nil
}

// LoadCredentials fetches stored credentials from disk, returning an empty struct when none exist.
func LoadCredentials() (Credentials, error) {
	file, err := credentialsFilePath()
	if err != nil {
		return Credentials{}, err
	}

	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return Credentials{}, nil
	}
	if err != nil {
		return Credentials{}, fmt.Errorf("read credentials file %s: %w", file, err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("parse credentials file %s: %w", file, err)
	}

	return creds, nil
}

// StoreCredentials writes the provided credentials to disk with user-only permissions.
func StoreCredentials(creds Credentials) error {
	file, err := credentialsFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return fmt.Errorf("create credentials directory: %w", err)
	}

	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(file, data, 0600); err != nil {
		return fmt.Errorf("write credentials file %s: %w", file, err)
	}

	return nil
}
