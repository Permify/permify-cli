// Package config holds config variables and functions
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// ClientCredentials stores endpoint and authentication credentials per profile in ~/.permify/credentials.
type ClientCredentials struct {
	PermifyURL  string `yaml:"permify_url,omitempty"`
	Token       string `yaml:"token,omitempty"`
	CertPath    string `yaml:"cert_path,omitempty"`
	CertKeyPath string `yaml:"cert_key_path,omitempty"`
}

// CoreConfig is the config struct
type CoreConfig struct {
	PermifyURL  string `yaml:"permify_url"`
	Tenant      string `yaml:"tenant"`
	Token       string `yaml:"token"`
	CertPath    string `yaml:"cert_path"`
	CertKeyPath string `yaml:"cert_key_path"`
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
	credentials, err := loadCredentials(profile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	cfg = applyCredentials(cfg, credentials)
	if cfg.PermifyURL == "" {
		return fmt.Errorf("permify url is empty for profile %s", profile)
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
	credentials, err := loadCredentials(profile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	CliConfig = applyCredentials(CliConfig, credentials)
	CliConfig.SslEnabled = usesTLS(CliConfig)
	return nil
}

// New initializes a new config file for permctl with the mentioned profile
func New(file string, profile string) error {
	profileConfigs.Profile = profile
	profileConfigs.File = file
	profileConfigs.Configs = make(map[string]CoreConfig)
	profileConfigs.Configs[profile] = sanitizeProfileConfig(CliConfig)
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
	profileConfigs.Configs[profile] = sanitizeProfileConfig(CliConfig)
	newConfigDataByte, err := yaml.Marshal(profileConfigs.Configs)
	if err != nil {
		return err
	}
	err = os.WriteFile(profileConfigs.File, newConfigDataByte, fs.FileMode(0644))
	if err != nil {
		return err
	}
	return writeCredentials(profile, extractCredentials(CliConfig))
}

func usesTLS(cfg CoreConfig) bool {
	return strings.HasPrefix(strings.ToLower(cfg.PermifyURL), "https://") || cfg.CertPath != ""
}

func credentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".permify", "credentials"), nil
}

func loadCredentials(profile string) (ClientCredentials, error) {
	credentialsFile, err := credentialsFilePath()
	if err != nil {
		return ClientCredentials{}, err
	}

	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return ClientCredentials{}, err
	}

	configs := map[string]ClientCredentials{}
	err = yaml.Unmarshal(data, &configs)
	if err != nil {
		logger.Log.Fatal("Error unmarshaling yaml")
	}

	return configs[profile], nil
}

func writeCredentials(profile string, credentials ClientCredentials) error {
	credentialsFile, err := credentialsFilePath()
	if err != nil {
		return err
	}

	configs := map[string]ClientCredentials{}
	data, err := os.ReadFile(credentialsFile)
	if err == nil {
		err = yaml.Unmarshal(data, &configs)
		if err != nil {
			logger.Log.Fatal("Error unmarshaling yaml")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	configs[profile] = credentials

	newCredentialsData, err := yaml.Marshal(configs)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(credentialsFile), fs.FileMode(0755))
	if err != nil {
		return err
	}

	return os.WriteFile(credentialsFile, newCredentialsData, fs.FileMode(0600))
}

func applyCredentials(cfg CoreConfig, credentials ClientCredentials) CoreConfig {
	if credentials.PermifyURL != "" {
		cfg.PermifyURL = credentials.PermifyURL
	}
	if credentials.Token != "" {
		cfg.Token = credentials.Token
	}
	if credentials.CertPath != "" {
		cfg.CertPath = credentials.CertPath
	}
	if credentials.CertKeyPath != "" {
		cfg.CertKeyPath = credentials.CertKeyPath
	}
	return cfg
}

func extractCredentials(cfg CoreConfig) ClientCredentials {
	return ClientCredentials{
		PermifyURL:  cfg.PermifyURL,
		Token:       cfg.Token,
		CertPath:    cfg.CertPath,
		CertKeyPath: cfg.CertKeyPath,
	}
}

func sanitizeProfileConfig(cfg CoreConfig) CoreConfig {
	cfg.PermifyURL = ""
	cfg.Token = ""
	cfg.CertPath = ""
	cfg.CertKeyPath = ""
	cfg.SslEnabled = false
	return cfg
}
