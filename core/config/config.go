// Package config holds config variables and functions
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment variable names for credential overrides
const (
	EnvPermifyEndpoint    = "PERMIFY_ENDPOINT"
	EnvPermifyURL         = "PERMIFY_URL"
	EnvPermifyToken       = "PERMIFY_TOKEN"
	EnvPermifyCertPath    = "PERMIFY_CERT_PATH"
	EnvPermifyCertKey     = "PERMIFY_CERT_KEY"
	EnvPermifyCertKeyPath = "PERMIFY_CERT_KEY_PATH"
)

// CliConfig is the global config variable
var CliConfig = CoreConfig{}

var profileConfigs = ProfileConfigs{}

// ProfileConfigs stores configs for all profiles
type ProfileConfigs struct {
	Configs map[string]CoreConfig
	File    string
	Profile string
}

// CoreConfig is the config struct
type CoreConfig struct {
	PermifyURL string `yaml:"permify_url"`
	Tenant     string `yaml:"tenant"`
	Token      string `yaml:"token,omitempty"`
	CertPath   string `yaml:"cert_path,omitempty"`
	CertKey    string `yaml:"cert_key,omitempty"`
	SslEnabled bool   `yaml:"-"`
}

// ComputeSslEnabled determines whether TLS should be enabled based on the endpoint scheme or certificate configuration
func ComputeSslEnabled(endpoint, certPath, certKey string) bool {
	ep := strings.ToLower(strings.TrimSpace(endpoint))
	if strings.HasPrefix(ep, "https://") || strings.HasPrefix(ep, "grpcs://") {
		return true
	}
	return strings.TrimSpace(certPath) != "" && strings.TrimSpace(certKey) != ""
}

// Validate checks that the required configuration and any TLS certificate files are valid
func (c CoreConfig) Validate() error {
	if strings.TrimSpace(c.PermifyURL) == "" {
		return errors.New("permify endpoint is missing")
	}

	certPath := strings.TrimSpace(c.CertPath)
	certKey := strings.TrimSpace(c.CertKey)
	hasCert := certPath != ""
	hasKey := certKey != ""

	if hasCert != hasKey {
		return errors.New("both cert_path and cert_key must be provided when configuring TLS")
	}

	if hasCert && hasKey {
		certInfo, err := os.Stat(certPath)
		if err != nil {
			return fmt.Errorf("certificate file not found at %s: %w", certPath, err)
		}
		if certInfo.IsDir() {
			return fmt.Errorf("certificate path %s is a directory", certPath)
		}

		keyInfo, err := os.Stat(certKey)
		if err != nil {
			return fmt.Errorf("certificate key file not found at %s: %w", certKey, err)
		}
		if keyInfo.IsDir() {
			return fmt.Errorf("certificate key path %s is a directory", certKey)
		}
	}

	return nil
}

// GetCredentials retrieves the configured endpoint, token, cert_path, and cert_key with environment variable overrides
func GetCredentials() (endpoint, token, certPath, certKey string) {
	endpoint = strings.TrimSpace(CliConfig.PermifyURL)
	token = strings.TrimSpace(CliConfig.Token)
	certPath = strings.TrimSpace(CliConfig.CertPath)
	certKey = strings.TrimSpace(CliConfig.CertKey)

	if envEndpoint := strings.TrimSpace(os.Getenv(EnvPermifyEndpoint)); envEndpoint != "" {
		endpoint = envEndpoint
	} else if envURL := strings.TrimSpace(os.Getenv(EnvPermifyURL)); envURL != "" {
		endpoint = envURL
	}

	if envToken := strings.TrimSpace(os.Getenv(EnvPermifyToken)); envToken != "" {
		token = envToken
	}

	if envCertPath := strings.TrimSpace(os.Getenv(EnvPermifyCertPath)); envCertPath != "" {
		certPath = envCertPath
	}

	if envCertKey := strings.TrimSpace(os.Getenv(EnvPermifyCertKey)); envCertKey != "" {
		certKey = envCertKey
	} else if envCertKeyPath := strings.TrimSpace(os.Getenv(EnvPermifyCertKeyPath)); envCertKeyPath != "" {
		certKey = envCertKeyPath
	}

	return endpoint, token, certPath, certKey
}

// IsConfigured checks if permctl cli has been configured
func IsConfigured(file string, profile string) error {
	_, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("config file %s does not exist", file)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var configs map[string]CoreConfig
	err = yaml.Unmarshal(data, &configs)
	if err != nil {
		return fmt.Errorf("error unmarshaling yaml: %w", err)
	}
	if configs == nil {
		configs = make(map[string]CoreConfig)
	}
	profileConfigs.Configs = configs
	cfg := profileConfigs.Configs[profile]
	if strings.TrimSpace(cfg.PermifyURL) == "" {
		return fmt.Errorf("permify url is empty for profile %s", profile)
	}
	if strings.TrimSpace(cfg.Tenant) == "" {
		return fmt.Errorf("tenant is empty for profile %s", profile)
	}
	return cfg.Validate()
}

// Load the permctl configuration specified by the user into the global variable
func Load(file string, profile string) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var configs map[string]CoreConfig
	err = yaml.Unmarshal(data, &configs)
	if err != nil {
		return fmt.Errorf("error unmarshaling yaml: %w", err)
	}
	if configs == nil {
		configs = make(map[string]CoreConfig)
	}
	profileConfigs.Configs = configs
	profileConfigs.File = file
	profileConfigs.Profile = profile
	CliConfig = profileConfigs.Configs[profile]
	CliConfig.SslEnabled = ComputeSslEnabled(CliConfig.PermifyURL, CliConfig.CertPath, CliConfig.CertKey)
	return nil
}

// New initializes a new config file for permctl with the mentioned profile
func New(file string, profile string) error {
	profileConfigs.Profile = profile
	profileConfigs.File = file
	profileConfigs.Configs = make(map[string]CoreConfig)
	CliConfig.SslEnabled = ComputeSslEnabled(CliConfig.PermifyURL, CliConfig.CertPath, CliConfig.CertKey)
	profileConfigs.Configs[profile] = CliConfig
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, newConfigDataByte, fs.FileMode(0600)); err != nil {
		return err
	}
	_ = os.Chmod(file, fs.FileMode(0600))
	return nil
}

// Write the config file
func Write() error {
	_, err := os.Stat(profileConfigs.File)
	if err != nil {
		return fmt.Errorf("%s config file does not exist", profileConfigs.File)
	}
	if profileConfigs.Configs == nil {
		profileConfigs.Configs = make(map[string]CoreConfig)
	}
	profile := profileConfigs.Profile
	CliConfig.SslEnabled = ComputeSslEnabled(CliConfig.PermifyURL, CliConfig.CertPath, CliConfig.CertKey)
	profileConfigs.Configs[profile] = CliConfig
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	if err := os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0600)); err != nil {
		return err
	}
	_ = os.Chmod(profileConfigs.File, fs.FileMode(0600))
	return nil
}
