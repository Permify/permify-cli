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
	PermifyURL  string `yaml:"permify_url,omitempty"`
	Tenant      string `yaml:"tenant"`
	Token       string `yaml:"-"`
	CertPath    string `yaml:"-"`
	CertKeyPath string `yaml:"-"`
	SslEnabled  bool   `yaml:"-"`
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
	cfg := profileConfigs.Configs[profile]
	if err = applyStoredCredentials(profile, &cfg); err != nil {
		return err
	}
	if cfg.PermifyURL == "" {
		return fmt.Errorf("permify url is empty for profile %s", profile)
	}
	if cfg.Tenant == "" {
		return fmt.Errorf("tenant is empty for profile %s", profile)
	}
	if (cfg.CertPath == "") != (cfg.CertKeyPath == "") {
		return fmt.Errorf("both cert_path and cert_key_path must be set for profile %s", profile)
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
	if err = applyStoredCredentials(profile, &CliConfig); err != nil {
		return err
	}
	CliConfig.SslEnabled = strings.HasPrefix(CliConfig.PermifyURL, "https")
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
	err = os.WriteFile(file, newConfigDataByte, fs.FileMode(0644))
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
	newConfigDataByte, err := yaml.Marshal(configsWithoutCredentials(profileConfigs.Configs))
	if err != nil {
		return err
	}
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0644))
	if err != nil {
		return err
	}
	return writeStoredCredentials(profileConfigs.Configs)
}

func configsWithoutCredentials(configs map[string]CoreConfig) map[string]CoreConfig {
	clean := make(map[string]CoreConfig, len(configs))
	for profile, cfg := range configs {
		cfg.PermifyURL = ""
		cfg.Token = ""
		cfg.CertPath = ""
		cfg.CertKeyPath = ""
		clean[profile] = cfg
	}
	return clean
}
