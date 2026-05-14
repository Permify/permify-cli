package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type storedCredentials struct {
	PermifyURL  string `yaml:"permify_url,omitempty"`
	Token       string `yaml:"token,omitempty"`
	CertPath    string `yaml:"cert_path,omitempty"`
	CertKeyPath string `yaml:"cert_key_path,omitempty"`
}

func credentialsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".permify", "credentials"), nil
}

func applyStoredCredentials(profile string, cfg *CoreConfig) error {
	credentials, ok, err := loadStoredCredentials(profile)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if credentials.PermifyURL != "" {
		cfg.PermifyURL = credentials.PermifyURL
	}
	cfg.Token = credentials.Token
	cfg.CertPath = credentials.CertPath
	cfg.CertKeyPath = credentials.CertKeyPath
	return nil
}

func loadStoredCredentials(profile string) (storedCredentials, bool, error) {
	credentialsFile, err := credentialsFilePath()
	if err != nil {
		return storedCredentials{}, false, err
	}
	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return storedCredentials{}, false, nil
		}
		return storedCredentials{}, false, err
	}
	if strings.TrimSpace(string(data)) == "" {
		return storedCredentials{}, false, nil
	}
	credentialsByProfile := map[string]storedCredentials{}
	if err = yaml.Unmarshal(data, &credentialsByProfile); err != nil {
		return storedCredentials{}, false, err
	}
	credentials, ok := credentialsByProfile[profile]
	return credentials, ok, nil
}

func writeStoredCredentials(configs map[string]CoreConfig) error {
	credentialsFile, err := credentialsFilePath()
	if err != nil {
		return err
	}
	credentialsByProfile := map[string]storedCredentials{}
	data, err := os.ReadFile(credentialsFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil && strings.TrimSpace(string(data)) != "" {
		if err = yaml.Unmarshal(data, &credentialsByProfile); err != nil {
			return err
		}
	}
	for profile, cfg := range configs {
		if cfg.PermifyURL == "" && cfg.Token == "" && cfg.CertPath == "" && cfg.CertKeyPath == "" {
			continue
		}
		credentialsByProfile[profile] = storedCredentials{
			PermifyURL:  cfg.PermifyURL,
			Token:       cfg.Token,
			CertPath:    cfg.CertPath,
			CertKeyPath: cfg.CertKeyPath,
		}
	}
	if err = os.MkdirAll(filepath.Dir(credentialsFile), 0700); err != nil {
		return err
	}
	newCredentialsData, err := yaml.Marshal(credentialsByProfile)
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsFile, newCredentialsData, 0600)
}
