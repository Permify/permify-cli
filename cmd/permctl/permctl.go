// Entry point for permify cli called permctl
package main

import (
	"os"
	"path/filepath"

	"github.com/Permify/permify-cli/core/cli"
)

func main() {
	home := os.Getenv("HOME")
	if home == "" {
		if derived, err := os.UserHomeDir(); err == nil {
			home = derived
		}
	}

	defaultConfig := filepath.Join(home, ".permctl")
	shortDescription := "permctl is a cli for managing and communicating with permify"
	permctl := cli.New("permctl", shortDescription, defaultConfig)
	cli.AddComponents(permctl.Cmd)
	permctl.Execute()
}
