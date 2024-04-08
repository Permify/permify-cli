// Package permission is cli sub command for communcating with permify permissions api
package permission

import (
	"github.com/Permify/permify-cli/utils"
	"github.com/spf13/cobra"
)

// New - Creates new permission command
func New() *cobra.Command {
	permissionCmd := &cobra.Command{
		Use:   "permission",
		Short: "use permify permissions api",
		Long:  "use permify permissions api",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help() 
		},
		PreRun: utils.CheckIfUnknownSubcommand,
	}
	checkCmd := CheckCmd{"check"}
	expandCmd := ExpandCmd{"expand"}
	lookupCmd := LookupCmd{"lookup"}
	subjectcmd := SubjectCmd{"subject"}

	permissionCmd.AddCommand(checkCmd.Cmd())
	permissionCmd.AddCommand(expandCmd.Cmd())
	permissionCmd.AddCommand(lookupCmd.Cmd())
	permissionCmd.AddCommand(subjectcmd.Cmd())

	return permissionCmd
}
