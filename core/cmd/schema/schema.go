// Package schema is cli sub command for communcating with permify schema api
package schema

import (
	"github.com/Permify/permify-cli/utils"
	"github.com/spf13/cobra"
)

// New - Creates new data command
func New() *cobra.Command {
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "use permify schema api",
		Long:  "use permify schema api",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help() 
		},
		PreRun: utils.CheckIfUnknownSubcommand,
	}
	readCmd := ReadCmd{"read"}
	writeCmd := WriteCmd{"write"}

	schemaCmd.AddCommand(readCmd.Cmd())
	schemaCmd.AddCommand(writeCmd.Cmd())

	return schemaCmd
}
