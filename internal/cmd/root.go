package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/internal/cmd/flags"
)

// NewRootCommand - Creates new root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permify-cli",
		Short: "command line interface for Permify",
		Long:  "command line interface for Permify",
	}

	flags.RegisterRootFlags(cmd)

	return cmd
}
