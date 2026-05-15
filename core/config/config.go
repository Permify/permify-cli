// Package config holds config variables and functions
package config

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

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
	PermifyURL string `yaml:"permify_url,omitempty"`
	Tenant     string `yaml:"tenant"`
	Token      string `yaml:"-"`
	CertPath   string `yaml:"-"`
	CertKey    string `yaml:"-"`
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
	if err = yaml.Unmarshal(data, &profileConfigs.Configs); err != nil {
		return fmt.Errorf("unmarshal config file %s: %w", file, err)
	}
	if profileConfigs.Configs == nil {
		profileConfigs.Configs = make(map[string]CoreConfig)
	}
	profileConfig := profileConfigs.Configs[profile]
	credentials, err := LoadStoredCredentials(profile)
	if err != nil {
		return err
	}
	if firstNonEmpty(credentials.EndpointValue(), profileConfig.PermifyURL) == "" {
		return fmt.Errorf("permify url is empty for profile %s", profile)
	}
	if profileConfig.Tenant == "" {
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
	if err = yaml.Unmarshal(data, &profileConfigs.Configs); err != nil {
		return fmt.Errorf("unmarshal config file %s: %w", file, err)
	}
	if profileConfigs.Configs == nil {
		profileConfigs.Configs = make(map[string]CoreConfig)
	}
	profileConfigs.File = file
	profileConfigs.Profile = profile
	CliConfig = profileConfigs.Configs[profile]
	if err = ApplyStoredCredentials(profile); err != nil {
		return err
	}
	CliConfig.SslEnabled = strings.HasPrefix(strings.ToLower(CliConfig.PermifyURL), "https") ||
		strings.HasPrefix(strings.ToLower(CliConfig.PermifyURL), "grpcs") ||
		CliConfig.CertPath != ""
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
	if profileConfigs.Configs == nil {
		profileConfigs.Configs = make(map[string]CoreConfig)
	}
	profileConfig := CliConfig
	profileConfig.PermifyURL = ""
	profileConfigs.Configs[profile] = profileConfig
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0644))
	return err
}
