//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTightenConfigFilePermissionsWithToken(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgs := map[string]CoreConfig{"default": {Token: "secret"}}
	if err := tightenConfigFilePermissions(path, cfgs); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if m := st.Mode().Perm(); m != 0o600 {
		t.Fatalf("want mode 0600, got %04o", m)
	}
}

func TestTightenConfigFilePermissionsNoToken(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgs := map[string]CoreConfig{"default": {PermifyURL: "localhost:1", Tenant: "t"}}
	if err := tightenConfigFilePermissions(path, cfgs); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if m := st.Mode().Perm(); m != 0o644 {
		t.Fatalf("want mode 0644, got %04o", m)
	}
}
