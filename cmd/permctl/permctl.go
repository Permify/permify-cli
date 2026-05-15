// Entry point for permify cli called permctl
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Permify/permify-cli/core/cli"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defaultConfig := filepath.Join(home, ".permctl")
	shortDescription := "permctl is a cli for managing and communicating with permify"
	permctl := cli.New("permctl", shortDescription, defaultConfig)
	cli.AddComponents(permctl.Cmd)
	permctl.Execute()
}
