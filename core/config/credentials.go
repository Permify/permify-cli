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

const credentialsFileMode fs.FileMode = 0o600

var credentialsPathOverride string

// StoredCredentials is the profile-scoped connection material persisted outside
// the normal tenant config file.
type StoredCredentials struct {
	Endpoint       string `yaml:"endpoint,omitempty"`
	LegacyURL      string `yaml:"permify_url,omitempty"`
	Token          string `yaml:"token,omitempty"`
	CertPath       string `yaml:"cert_path,omitempty"`
	CertKey        string `yaml:"cert_key,omitempty"`
	LegacyCertKey string `yaml:"cert_key_path,omitempty"`
}

func (c StoredCredentials) EndpointValue() string {
	return firstNonEmpty(c.Endpoint, c.LegacyURL)
}

func (c StoredCredentials) CertKeyValue() string {
	return firstNonEmpty(c.CertKey, c.LegacyCertKey)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func credentialsFilePathForHome(home string) string {
	return filepath.Join(home, ".permify", "credentials")
}

func credentialsFilePath() (string, error) {
	if credentialsPathOverride != "" {
		return credentialsPathOverride, nil
	}

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

func loadStoredCredentialsFile(path string) (map[string]StoredCredentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]StoredCredentials{}, nil
		}
		return nil, err
	}

	var profiles map[string]StoredCredentials
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("unmarshal credentials file %s: %w", path, err)
	}
	if profiles == nil {
		profiles = map[string]StoredCredentials{}
	}
	return profiles, nil
}

func writeStoredCredentialsFile(path string, profiles map[string]StoredCredentials) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	for profile, credentials := range profiles {
		credentials.Endpoint = credentials.EndpointValue()
		credentials.LegacyURL = ""
		credentials.CertKey = credentials.CertKeyValue()
		credentials.LegacyCertKey = ""
		profiles[profile] = credentials
	}

	data, err := yaml.Marshal(profiles)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return os.WriteFile(path, data, credentialsFileMode)
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".credentials-*.tmp")
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

	if err := file.Chmod(credentialsFileMode); err != nil {
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
	path, err := credentialsFilePath()
	if err != nil {
		return StoredCredentials{}, err
	}
	profiles, err := loadStoredCredentialsFile(path)
	if err != nil {
		return StoredCredentials{}, err
	}
	return profiles[profile], nil
}

func ApplyStoredCredentials(profile string) error {
	credentials, err := LoadStoredCredentials(profile)
	if err != nil {
		return err
	}

	CliConfig.PermifyURL = firstNonEmpty(credentials.EndpointValue(), CliConfig.PermifyURL)
	CliConfig.Token = strings.TrimSpace(credentials.Token)
	CliConfig.CertPath = strings.TrimSpace(credentials.CertPath)
	CliConfig.CertKey = strings.TrimSpace(credentials.CertKeyValue())
	return nil
}

func WriteStoredCredentials(profile string, cfg CoreConfig) error {
	path, err := credentialsFilePath()
	if err != nil {
		return err
	}

	profiles, err := loadStoredCredentialsFile(path)
	if err != nil {
		return err
	}
	profiles[profile] = StoredCredentials{
		Endpoint: strings.TrimSpace(cfg.PermifyURL),
		Token:    strings.TrimSpace(cfg.Token),
		CertPath: strings.TrimSpace(cfg.CertPath),
		CertKey:  strings.TrimSpace(cfg.CertKey),
	}
	return writeStoredCredentialsFile(path, profiles)
}
