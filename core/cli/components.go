package cli

import (
	"github.com/Permify/permify-cli/core/cmd/data"
	"github.com/Permify/permify-cli/core/cmd/permission"
	"github.com/Permify/permify-cli/core/cmd/schema"
	"github.com/Permify/permify-cli/core/cmd/tenancy"
	"github.com/spf13/cobra"
)


func AddComponents(rootCmd *cobra.Command) {
	permissionCmd := permission.New()
	tenancyCmd := tenancy.New()
	dataCmd := data.New()
	schemaCmd := schema.New()

	rootCmd.AddCommand(permissionCmd)
	rootCmd.AddCommand(tenancyCmd)
	rootCmd.AddCommand(dataCmd)
	rootCmd.AddCommand(schemaCmd)
}
