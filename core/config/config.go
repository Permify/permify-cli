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

// GetStoredCredentials retrieves and validates the stored credentials from CliConfig.
func GetStoredCredentials(endpointOverride string) (CoreConfig, error) {
	cfg := CliConfig
	if strings.TrimSpace(endpointOverride) != "" {
		cfg.PermifyURL = strings.TrimSpace(endpointOverride)
	}
	if strings.TrimSpace(cfg.PermifyURL) == "" {
		return cfg, errors.New("permify endpoint is missing from configuration")
	}
	if cfg.CertKey != "" && cfg.CertPath == "" {
		return cfg, errors.New("cert_path must be provided when cert_key is configured")
	}
	if cfg.CertPath != "" {
		if _, err := os.Stat(cfg.CertPath); err != nil {
			return cfg, fmt.Errorf("certificate file %s is not accessible: %w", cfg.CertPath, err)
		}
	}
	if cfg.CertKey != "" {
		if _, err := os.Stat(cfg.CertKey); err != nil {
			return cfg, fmt.Errorf("certificate key file %s is not accessible: %w", cfg.CertKey, err)
		}
	}
	cfg.SslEnabled = strings.HasPrefix(cfg.PermifyURL, "https") || cfg.CertPath != ""
	return cfg, nil
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
	if profileConfigs.Configs[profile].PermifyURL == "" {
		return fmt.Errorf("permify url is empty for profile %s", profile)
	}
	if profileConfigs.Configs[profile].Tenant == "" {
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
