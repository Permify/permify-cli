package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Credentials are stored separately from the main profile YAML config to avoid
// repeatedly prompting users for connection details.
type Credentials struct {
	Endpoint string `yaml:"endpoint,omitempty"`
	Token    string `yaml:"token,omitempty"`
	CertPath string `yaml:"cert_path,omitempty"`
	CertKey  string `yaml:"cert_key,omitempty"`
}

func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".permify", "credentials"), nil
	}
	return filepath.Join(home, ".permify", "credentials"), nil
}

func LoadCredentials() (Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return Credentials{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Credentials{}, nil
		}
		return Credentials{}, err
	}
	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("unmarshal credentials: %w", err)
	}
	return creds, nil
}

func WriteCredentials(creds Credentials) (string, error) {
	path, err := credentialsPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	data, err := yaml.Marshal(creds)
	if err != nil {
		return "", err
	}
	// Credentials may include tokens; restrict permissions.
	if err := os.WriteFile(path, data, fs.FileMode(0o600)); err != nil {
		return "", err
	}
	return path, nil
}

