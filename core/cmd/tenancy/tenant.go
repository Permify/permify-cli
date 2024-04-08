// Package tenancy is cli sub command for communcating with permify tenancy api
package tenancy

import (
	"github.com/Permify/permify-cli/utils"
	"github.com/spf13/cobra"
)

// New - Creates new tenant command
func New() *cobra.Command {
	tenancyCmd := &cobra.Command{
		Use:   "tenant",
		Short: "use permify tenants api",
		Long:  "use permify tenants api",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help() 
		},
		PreRun: utils.CheckIfUnknownSubcommand,
	}
	createCmd := CreateCmd{"create"}
	deleteCmd := DeleteCmd{"delete"}
	listCmd := ListCmd{"list"}

	tenancyCmd.AddCommand(createCmd.Cmd())
	tenancyCmd.AddCommand(deleteCmd.Cmd())
	tenancyCmd.AddCommand(listCmd.Cmd())

	return tenancyCmd
}
