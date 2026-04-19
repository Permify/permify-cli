// Package config holds config variables and functions
package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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
	PermifyURL  string `yaml:"permify_url"`
	Tenant      string `yaml:"tenant"`
	Token       string `yaml:"token,omitempty"`
	CertPath    string `yaml:"cert_path,omitempty"`
	CertKeyPath string `yaml:"cert_key_path,omitempty"`
	SslEnabled  bool   `yaml:"-"`
}

func anyProfileStoresToken(configs map[string]CoreConfig) bool {
	for _, c := range configs {
		if strings.TrimSpace(c.Token) != "" {
			return true
		}
	}
	return false
}

// writeConfigFile writes YAML bytes and sets permissions without leaving a token-bearing file
// world-readable between truncate and chmod (Unix: temp file + rename with mode 0600).
func writeConfigFile(path string, data []byte, configs map[string]CoreConfig) error {
	sensitive := anyProfileStoresToken(configs)
	if runtime.GOOS == "windows" {
		perm := fs.FileMode(0o644)
		if sensitive {
			perm = 0o666
		}
		return os.WriteFile(path, data, perm)
	}
	if sensitive {
		return atomicWriteFileUnix(path, data, 0o600)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return os.Chmod(path, 0o644)
}

func atomicWriteFileUnix(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".permctl-config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := f.Name()
	cleanup := true
	defer func() {
		_ = f.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()
	if err := f.Chmod(perm); err != nil {
		return fmt.Errorf("chmod temp config: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync temp config: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	cleanup = false
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
	return writeConfigFile(file, newConfigDataByte, profileConfigs.Configs)
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
	return writeConfigFile(profileConfigs.File, newConfigDataByte, profileConfigs.Configs)
}
