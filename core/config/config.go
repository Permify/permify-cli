// Package config holds config variables and functions
package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Permify/permify-cli/core/logger"
	"gopkg.in/yaml.v3"
)

// CliConfig is the global config variable.
var CliConfig = CoreConfig{}

// CliCredentials stores the connection secrets for the active profile.
var CliCredentials = ConnectionCredentials{}

var profileConfigs = ProfileConfigs{}
var credentialsConfigs = CredentialProfiles{}

// ProfileConfigs stores configs for all profiles.
type ProfileConfigs struct {
	Configs map[string]CoreConfig
	File    string
	Profile string
}

// CredentialProfiles stores credentials for all profiles.
type CredentialProfiles struct {
	Configs map[string]ConnectionCredentials
	File    string
	Profile string
}

// CoreConfig is the config struct.
type CoreConfig struct {
	PermifyURL string `yaml:"permify_url"`
	Tenant     string `yaml:"tenant"`
	SslEnabled bool   `yaml:"-"`
}

// ConnectionCredentials keeps endpoint and connection secrets outside the main config file.
type ConnectionCredentials struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
	CertPath string `yaml:"cert_path"`
	CertKey  string `yaml:"cert_key"`
}

func defaultCredentialsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return filepath.Join(".", ".permify", "credentials")
	}
	return filepath.Join(homeDir, ".permify", "credentials")
}

func ensureCredentialState(profile string) {
	if credentialsConfigs.Configs == nil {
		credentialsConfigs.Configs = make(map[string]ConnectionCredentials)
	}
	credentialsConfigs.Profile = profile
}

func loadCredentials(file string, profile string) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	ensureCredentialState(profile)
	credentialsConfigs.File = file
	err = yaml.Unmarshal(data, &credentialsConfigs.Configs)
	if err != nil {
		logger.Log.Fatal("Error unmarshaling yaml")
	}
	CliCredentials = credentialsConfigs.Configs[profile]
	if CliCredentials.Endpoint != "" {
		CliConfig.PermifyURL = CliCredentials.Endpoint
		CliConfig.SslEnabled = strings.HasPrefix(CliCredentials.Endpoint, "https")
	}
	return nil
}

// LoadCredentials loads the credentials file into the global variable.
func LoadCredentials(file string, profile string) error {
	return loadCredentials(file, profile)
}

// IsConfigured checks if permctl cli has been configured.
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

// Load the permctl configuration specified by the user into the global variable.
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
	CliCredentials = ConnectionCredentials{}
	_ = loadCredentials(defaultCredentialsPath(), profile)
	if CliCredentials.Endpoint != "" {
		CliConfig.PermifyURL = CliCredentials.Endpoint
		CliConfig.SslEnabled = strings.HasPrefix(CliCredentials.Endpoint, "https")
	}
	return err
}

// New initializes a new config file for permctl with the mentioned profile.
func New(file string, profile string) error {
	profileConfigs.Profile = profile
	profileConfigs.File = file
	profileConfigs.Configs = make(map[string]CoreConfig)
	profileConfigs.Configs[profile] = CliConfig
	ensureCredentialState(profile)
	credentialsConfigs.File = defaultCredentialsPath()
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	err = os.WriteFile(file, newConfigDataByte, fs.FileMode(0644))
	return err
}

// Write the config file.
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
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0644))
	return err
}

// WriteCredentials writes the credentials file using the active profile.
func WriteCredentials() error {
	if credentialsConfigs.File == "" {
		credentialsConfigs.File = defaultCredentialsPath()
	}
	ensureCredentialState(credentialsConfigs.Profile)
	if err := os.MkdirAll(filepath.Dir(credentialsConfigs.File), 0700); err != nil {
		return err
	}
	credentialsConfigs.Configs[credentialsConfigs.Profile] = CliCredentials
	newCredentialsDataByte, err := yaml.Marshal(credentialsConfigs.Configs)
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsConfigs.File, newCredentialsDataByte, fs.FileMode(0600))
}
