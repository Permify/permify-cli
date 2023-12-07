package main

import (
	"os"

	"github.com/Permify/permify-cli/internal/cmd"
	"github.com/Permify/permify-cli/internal/cmd/permission"
)

func main() {
	rootCmd := cmd.NewRootCommand()

	permissionCmd := permission.NewPermissionCommand()

	checkCmd := permission.NewCheckCommand()
	permissionCmd.AddCommand(checkCmd)

	rootCmd.AddCommand(permissionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
