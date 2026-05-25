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
	PermifyURL string `yaml:"permify_url"`
	Tenant     string `yaml:"tenant"`
	Token      string `yaml:"token,omitempty"`
	CertPath   string `yaml:"cert_path,omitempty"`
	CertKey    string `yaml:"cert_key,omitempty"`
	SslEnabled bool   `yaml:"-"`
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
	err = writeConfigFile(file, newConfigDataByte)
	return err
}

func writeConfigFile(file string, data []byte) error {
	if err := os.WriteFile(file, data, fs.FileMode(0600)); err != nil {
		return err
	}
	return os.Chmod(file, fs.FileMode(0600))
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
	err = writeConfigFile(profileConfigs.File, newConfigDataByte)
	return err
}
