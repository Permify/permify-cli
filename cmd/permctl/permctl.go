// Entry point for permify cli called permctl
package main

import (
	"os"
	"path/filepath"

	"github.com/Permify/permify-cli/core/cli"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	defaultConfig := filepath.Join(home, ".permctl")
	shortDescription := "permctl is a cli for managing and communicating with permify"
	permctl := cli.New("permctl", shortDescription, defaultConfig)
	cli.AddComponents(permctl.Cmd)
	permctl.Execute()
}
