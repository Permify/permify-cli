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
	Configs map[string]CoreConfig
	File    string
	Profile string
}

// CoreConfig is the config struct
type CoreConfig struct {
	PermifyURL   string `yaml:"permify_url"`
	Tenant       string `yaml:"tenant"`
	Token        string `yaml:"token,omitempty"`
	CertPath     string `yaml:"cert_path,omitempty"`
	CertKeyPath  string `yaml:"cert_key_path,omitempty"`
	SslEnabled   bool   `yaml:"-"`
}

const (
	// Env overrides (do not get written back to config automatically)
	PermifyURLEnv     = "PERMCTL_PERMIFY_URL"
	PermifyTenantEnv  = "PERMCTL_TENANT"
	PermifyTokenEnv   = "PERMCTL_TOKEN"
	PermifyCertEnv    = "PERMCTL_CERT_PATH"
	PermifyCertKeyEnv = "PERMCTL_CERT_KEY_PATH"
)

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
	applyEnvOverrides(&CliConfig)
	CliConfig.SslEnabled = strings.HasPrefix(CliConfig.PermifyURL, "https")
	return err
}

func applyEnvOverrides(cfg *CoreConfig) {
	if v := strings.TrimSpace(os.Getenv(PermifyURLEnv)); v != "" {
		cfg.PermifyURL = v
	}
	if v := strings.TrimSpace(os.Getenv(PermifyTenantEnv)); v != "" {
		cfg.Tenant = v
	}
	if v := strings.TrimSpace(os.Getenv(PermifyTokenEnv)); v != "" {
		cfg.Token = v
	}
	if v := strings.TrimSpace(os.Getenv(PermifyCertEnv)); v != "" {
		cfg.CertPath = v
	}
	if v := strings.TrimSpace(os.Getenv(PermifyCertKeyEnv)); v != "" {
		cfg.CertKeyPath = v
	}
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
	// config may contain secrets (token), so default to owner-read/write.
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
	// config may contain secrets (token), so default to owner-read/write.
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0600))
	return err
}
