// Package config holds config variables and functions
package config

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/Permify/permify-cli/core/logger"
	"gopkg.in/yaml.v3"
)

// CliConfig is the global config variable
var CliConfig = CoreConfig{}

var profileConfigs = ProfileConfigs{}

// ProfileConfigs stores configs for all profiles
type ProfileConfigs struct {
	Configs map[string]CoreConfig `yaml:",inline"`
	File    string                `yaml:"-"`
	Profile string                `yaml:"-"`
}

// CoreConfig is the config struct storing connection parameters and credentials
type CoreConfig struct {
	PermifyURL string `yaml:"permify_url,omitempty"`
	Endpoint   string `yaml:"endpoint,omitempty"`
	Tenant     string `yaml:"tenant"`
	Token      string `yaml:"token,omitempty"`
	CertPath   string `yaml:"cert_path,omitempty"`
	CertKey    string `yaml:"cert_key,omitempty"`
	Insecure   bool   `yaml:"insecure,omitempty"`
	SslEnabled bool   `yaml:"-"`
}

// GetEndpoint returns the resolved endpoint URL (checking Endpoint then PermifyURL)
func (c *CoreConfig) GetEndpoint() string {
	if c.Endpoint != "" {
		return c.Endpoint
	}
	return c.PermifyURL
}

// SetEndpoint sets both Endpoint and PermifyURL for compatibility
func (c *CoreConfig) SetEndpoint(endpoint string) {
	c.Endpoint = endpoint
	c.PermifyURL = endpoint
}

// ValidateCredentials verifies that configured credential files exist and are accessible
func (c *CoreConfig) ValidateCredentials() error {
	if c.CertPath != "" {
		if _, err := os.Stat(c.CertPath); err != nil {
			return fmt.Errorf("cert_path '%s' is not accessible: %w", c.CertPath, err)
		}
	}
	if c.CertKey != "" {
		if _, err := os.Stat(c.CertKey); err != nil {
			return fmt.Errorf("cert_key '%s' is not accessible: %w", c.CertKey, err)
		}
	}
	return nil
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
	err = yaml.Unmarshal(data, &profileConfigs.Configs)
	if err != nil {
		logger.Log.Fatal("Error unmarshaling yaml")
	}
	cfg, ok := profileConfigs.Configs[profile]
	if !ok {
		return fmt.Errorf("profile %s not found in config", profile)
	}
	if cfg.PermifyURL == "" && cfg.Endpoint == "" {
		return fmt.Errorf("permify url / endpoint is empty for profile %s", profile)
	}
	if cfg.Tenant == "" {
		return fmt.Errorf("tenant is empty for profile %s", profile)
	}
	return nil
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
	err = yaml.Unmarshal(data, &profileConfigs.Configs)
	if err != nil {
		logger.Log.Fatal("Error unmarshaling yaml")
	}
	profileConfigs.File = file
	profileConfigs.Profile = profile
	CliConfig = profileConfigs.Configs[profile]
	if CliConfig.PermifyURL == "" && CliConfig.Endpoint != "" {
		CliConfig.PermifyURL = CliConfig.Endpoint
	} else if CliConfig.Endpoint == "" && CliConfig.PermifyURL != "" {
		CliConfig.Endpoint = CliConfig.PermifyURL
	}
	CliConfig.SslEnabled = strings.HasPrefix(CliConfig.PermifyURL, "https") || CliConfig.CertPath != ""
	return err
}

// New initializes a new config file for permctl with the mentioned profile
func New(file string, profile string) error {
	profileConfigs.Profile = profile
	profileConfigs.File = file
	profileConfigs.Configs = make(map[string]CoreConfig)
	profileConfigs.Configs[profile] = CliConfig
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	// Store with restricted 0600 permissions to protect tokens and private keys
	err = os.WriteFile(file, newConfigDataByte, fs.FileMode(0600))
	return err
}

// Write the config file
func Write() error {
	_, err := os.Stat(profileConfigs.File)
	if err != nil {
		return fmt.Errorf("%s config file does not exist", profileConfigs.File)
	}
	profile := profileConfigs.Profile
	profileConfigs.Configs[profile] = CliConfig
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	// Store with restricted 0600 permissions to protect tokens and private keys
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0600))
	return err
}
