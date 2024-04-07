// Package data is cli sub command for communcating with permify data api
package data

import (
	"github.com/Permify/permify-cli/utils"
	"github.com/spf13/cobra"
)

// New - Creates new data command
func New() *cobra.Command {
	dataCmd := &cobra.Command{
		Use:   "data",
		Short: "use permify data api",
		Long:  "use permify data api",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help() 
		},
		PreRun: utils.CheckIfUnknownSubcommand,
	}
	writeCmd := WriteCmd{"write"}
	readCmd := ReadCmd{"read"}

	dataCmd.AddCommand(writeCmd.Cmd())
	dataCmd.AddCommand(readCmd.Cmd())

	return dataCmd
}
