// Package config holds config variables and functions
package config

import (
	"errors"
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
	Configs map[string]CoreConfig
	File    string
	Profile string
}

// CoreConfig is the config struct storing endpoint, tenant, token, cert path, and cert key
type CoreConfig struct {
	PermifyURL string `yaml:"permify_url"`
	Tenant     string `yaml:"tenant"`
	Token      string `yaml:"token,omitempty"`
	CertPath   string `yaml:"cert_path,omitempty"`
	CertKey    string `yaml:"cert_key,omitempty"`
	SslEnabled bool   `yaml:"-"`
}

// Validate checks if the stored credentials and certificate paths are valid
func (c CoreConfig) Validate() error {
	if strings.TrimSpace(c.PermifyURL) == "" {
		return errors.New("permify url is empty")
	}
	if c.CertKey != "" && c.CertPath == "" {
		return errors.New("cert_path is required when cert_key is provided")
	}
	if c.CertPath != "" {
		if _, err := os.Stat(c.CertPath); err != nil {
			return fmt.Errorf("cert_path file %s does not exist or is not accessible: %w", c.CertPath, err)
		}
	}
	if c.CertKey != "" {
		if _, err := os.Stat(c.CertKey); err != nil {
			return fmt.Errorf("cert_key file %s does not exist or is not accessible: %w", c.CertKey, err)
		}
	}
	return nil
}

// GetCredentials retrieves the active profile's stored credentials from CliConfig
func GetCredentials() (CoreConfig, error) {
	if err := CliConfig.Validate(); err != nil {
		return CoreConfig{}, err
	}
	return CliConfig, nil
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
	if !ok || cfg.PermifyURL == "" {
		return fmt.Errorf("permify url is empty for profile %s", profile)
	}
	if cfg.Tenant == "" {
		return fmt.Errorf("tenant is empty for profile %s", profile)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid credentials for profile %s: %w", profile, err)
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
		return fmt.Errorf("error unmarshaling yaml: %w", err)
	}
	profileConfigs.File = file
	profileConfigs.Profile = profile
	CliConfig = profileConfigs.Configs[profile]
	CliConfig.SslEnabled = strings.HasPrefix(CliConfig.PermifyURL, "https") || CliConfig.CertPath != ""
	return nil
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
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0600))
	return err
}
