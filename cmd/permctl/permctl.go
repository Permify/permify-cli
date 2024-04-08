// Entry point for permify cli called permctl
package main

import (
	"fmt"
	"os"

	"github.com/Permify/permify-cli/core/cli"
)

func main() {
	home := os.Getenv("HOME")
	defaultConfig := fmt.Sprintf("%s/.permctl", home)
	shortDescription := "permctl is a cli for managing and communicating with permify"
	permctl	:= cli.New("permctl", shortDescription, defaultConfig)	
	cli.AddComponents(permctl.Cmd)
	permctl.Execute()
}
