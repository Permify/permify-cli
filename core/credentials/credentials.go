// Package credentials handles secure storage and retrieval of permctl
// credentials (endpoint, token, mTLS cert/key paths). Credentials are
// persisted as YAML on the local filesystem with 0600 permissions so that
// only the owning user can read/write them.
//
// Storage layout:
//
//	$PERMIFY_CREDENTIALS_FILE   (when set, takes precedence)
//	$HOME/.permctl-credentials.yaml   (default)
//
// The file groups credentials under the active profile name, mirroring the
// existing config layout in core/config:
//
//	default:
//	  endpoint: localhost:3478
//	  token: secret-bearer-token
//	  cert_path: /etc/permify/client.crt
//	  cert_key:  /etc/permify/client.key
//	staging:
//	  endpoint: staging.permify.example:443
//	  ...
package credentials

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// DefaultProfile is the profile used when callers do not specify one.
const DefaultProfile = "default"

// defaultFileName is the filename used under $HOME when no override is set.
const defaultFileName = ".permctl-credentials.yaml"

// envFileOverride lets users (and tests) point at a non-default file.
const envFileOverride = "PERMIFY_CREDENTIALS_FILE"

// secureFileMode is the permission mode for the credentials file. 0600 means
// the owning user is the only one able to read or write the file.
const secureFileMode fs.FileMode = 0o600

// secureDirMode is the permission mode used when creating the credentials
// directory.
const secureDirMode fs.FileMode = 0o700

// ErrNotFound is returned when no credentials are stored at the requested
// location, or when no entry exists for the requested profile. Callers can
// use errors.Is to detect this case and fall back to flags / interactive
// prompts.
var ErrNotFound = errors.New("credentials not found")

// Credentials holds the connection material required to talk to a Permify
// server. All fields are optional individually; consumers decide which ones
// they need (for example, a plaintext endpoint needs no cert paths).
type Credentials struct {
	Endpoint string `yaml:"endpoint,omitempty"`
	Token    string `yaml:"token,omitempty"`
	CertPath string `yaml:"cert_path,omitempty"`
	CertKey  string `yaml:"cert_key,omitempty"`
}

// IsZero reports whether c carries no information at all. Useful for
// distinguishing "stored but empty" from a real value.
func (c Credentials) IsZero() bool {
	return c.Endpoint == "" && c.Token == "" && c.CertPath == "" && c.CertKey == ""
}

// store is the on-disk representation: a map of profile name -> credentials.
type store map[string]Credentials

// GetCredentialsPath returns the absolute path to the credentials file.
// The lookup order is:
//
//  1. $PERMIFY_CREDENTIALS_FILE if set
//  2. $HOME/.permctl-credentials.yaml otherwise
//
// On Windows, where $HOME is not always set, os.UserHomeDir is used as a
// fallback so the function works out of the box on all supported platforms.
func GetCredentialsPath() (string, error) {
	if override := os.Getenv(envFileOverride); override != "" {
		return override, nil
	}
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, defaultFileName), nil
}

// userHomeDir resolves the user's home directory, preferring $HOME for
// parity with the existing config package and falling back to
// os.UserHomeDir when unset (the typical Windows case).
func userHomeDir() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	return os.UserHomeDir()
}

// Load reads credentials for the given profile from the default file. If
// profile is empty, DefaultProfile is used. Returns ErrNotFound when the
// file does not exist or the profile has no entry.
func Load(profile string) (*Credentials, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path, profile)
}

// LoadFrom reads credentials from an explicit path. Exposed primarily for
// testing but also useful for callers that want a non-default location
// without touching the environment.
func LoadFrom(path, profile string) (*Credentials, error) {
	if profile == "" {
		profile = DefaultProfile
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return nil, fmt.Errorf("read credentials file: %w", err)
	}
	// Best-effort permission audit: warn (via error) on Unix if the file is
	// world/group readable. We do not auto-fix to avoid surprising the user.
	if err := assertSecurePermissions(path); err != nil {
		return nil, err
	}
	s := store{}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("parse credentials file %s: %w", path, err)
		}
	}
	creds, ok := s[profile]
	if !ok {
		return nil, fmt.Errorf("%w: profile %q in %s", ErrNotFound, profile, path)
	}
	return &creds, nil
}

// Save writes credentials for the given profile to the default file,
// preserving entries for other profiles. The file is created with 0600
// permissions; existing files have their permissions tightened to 0600 if
// they were previously broader.
func Save(profile string, creds *Credentials) error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}
	return SaveTo(path, profile, creds)
}

// SaveTo writes credentials to an explicit path. See Save for behavior.
func SaveTo(path, profile string, creds *Credentials) error {
	if creds == nil {
		return errors.New("credentials must not be nil")
	}
	if profile == "" {
		profile = DefaultProfile
	}

	// Load the existing file (if any) so other profiles are preserved.
	s := store{}
	if data, err := os.ReadFile(path); err == nil {
		if len(data) > 0 {
			if err := yaml.Unmarshal(data, &s); err != nil {
				return fmt.Errorf("parse existing credentials file %s: %w", path, err)
			}
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read credentials file: %w", err)
	}

	s[profile] = *creds
	return writeStore(path, s)
}

// Delete removes the entry for the given profile. If no other profiles
// remain, the file itself is removed. Returns nil if the file or profile
// already does not exist.
func Delete(profile string) error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}
	return DeleteFrom(path, profile)
}

// DeleteFrom removes the entry for the given profile from an explicit path.
func DeleteFrom(path, profile string) error {
	if profile == "" {
		profile = DefaultProfile
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read credentials file: %w", err)
	}
	s := store{}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("parse credentials file %s: %w", path, err)
		}
	}
	if _, ok := s[profile]; !ok {
		return nil
	}
	delete(s, profile)
	if len(s) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove empty credentials file: %w", err)
		}
		return nil
	}
	return writeStore(path, s)
}

// writeStore serializes the entire credential map to disk atomically with
// 0600 permissions. Used by SaveTo and DeleteFrom.
func writeStore(path string, s store) error {
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, secureDirMode); err != nil {
			return fmt.Errorf("create credentials directory: %w", err)
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".permctl-credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp credentials file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(out); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp credentials file: %w", err)
	}
	if err := tmp.Chmod(secureFileMode); err != nil && runtime.GOOS != "windows" {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp credentials file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp credentials file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("install credentials file: %w", err)
	}
	if err := os.Chmod(path, secureFileMode); err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("chmod credentials file: %w", err)
	}
	return nil
}

// assertSecurePermissions returns a descriptive error when the credentials
// file is more permissive than 0600 on POSIX systems. Windows ACL semantics
// don't map cleanly to mode bits, so the check is skipped there.
func assertSecurePermissions(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat credentials file: %w", err)
	}
	// Reject any bits beyond owner read/write.
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		return fmt.Errorf(
			"insecure permissions on %s: have %#o, expected %#o (run: chmod 600 %s)",
			path, mode, secureFileMode, path,
		)
	}
	return nil
}
