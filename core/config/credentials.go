package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type StoredCredentials struct {
	PermifyURL  string `yaml:"permify_url,omitempty"`
	Token       string `yaml:"token,omitempty"`
	CertPath    string `yaml:"cert_path,omitempty"`
	CertKeyPath string `yaml:"cert_key_path,omitempty"`
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func credentialsFilePathForHome(home string) string {
	return filepath.Join(home, ".permify", "credentials")
}

func credentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve credentials path: %w", err)
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return "", fmt.Errorf("resolve credentials path: empty home directory")
	}
	return credentialsFilePathForHome(home), nil
}

func loadStoredCredentialsFile() (map[string]StoredCredentials, error) {
	path, err := credentialsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]StoredCredentials{}, nil
		}
		return nil, err
	}
	var profiles map[string]StoredCredentials
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}
	if profiles == nil {
		profiles = map[string]StoredCredentials{}
	}
	return profiles, nil
}

func writeStoredCredentialsFile(path string, data []byte) error {
	if runtime.GOOS == "windows" {
		return os.WriteFile(path, data, fs.FileMode(0o600))
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".permify-credentials-*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	cleanup := true
	defer func() {
		_ = file.Close()
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	if err := file.Chmod(0o600); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func LoadStoredCredentials(profile string) (StoredCredentials, error) {
	credentials, err := loadStoredCredentialsFile()
	if err != nil {
		return StoredCredentials{}, err
	}
	return credentials[profile], nil
}

func ApplyStoredCredentials(profile string) error {
	storedCredentials, err := LoadStoredCredentials(profile)
	if err != nil {
		return err
	}

	CliConfig.PermifyURL = firstNonEmpty(storedCredentials.PermifyURL, CliConfig.PermifyURL)
	CliConfig.Token = storedCredentials.Token
	CliConfig.CertPath = storedCredentials.CertPath
	CliConfig.CertKeyPath = storedCredentials.CertKeyPath
	return nil
}

func WriteStoredCredentials(profile string, cfg CoreConfig) error {
	credentials, err := loadStoredCredentialsFile()
	if err != nil {
		return err
	}
	if credentials == nil {
		credentials = map[string]StoredCredentials{}
	}

	path, err := credentialsFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	credentials[profile] = StoredCredentials{
		PermifyURL:  cfg.PermifyURL,
		Token:       cfg.Token,
		CertPath:    cfg.CertPath,
		CertKeyPath: cfg.CertKeyPath,
	}

	data, err := yaml.Marshal(credentials)
	if err != nil {
		return err
	}
	return writeStoredCredentialsFile(path, data)
}
