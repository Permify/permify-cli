package permission

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/core/config"
	"github.com/Permify/permify-cli/tui"
	"github.com/Permify/permify-cli/utils"
	v1 "github.com/Permify/permify-go/generated/base/v1"
)

// ExpandCmd - implements permission check api
type ExpandCmd struct {
	Command string
}

// Cmd - check command
func (ec *ExpandCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ec.Command,
		Short: "run expand request",
		Run:  ec.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("entity", "e", "", "entity identifier specified as - <type>:<id>")
	cmd.Flags().StringP("permission", "p", "", "[Optional] permission to check")
	return cmd
}

func (ec *ExpandCmd) Run(cmd *cobra.Command, args []string) {
	entity, _ := cmd.Flags().GetString("entity")
	if entity == "" {
		newEntity, err := tui.StringPrompt("Enter entity string", "<type>:<id>", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		entity = newEntity
	}
	parsedEntity, err := utils.ParseEntity(entity)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	permission, _ := cmd.Flags().GetString("permission")
	schemaVersion, _ := cmd.Flags().GetString("schema")

	permissionClient := Client()
	expandRequest := &v1.PermissionExpandRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.PermissionExpandRequestMetadata{
			SchemaVersion: schemaVersion,
		},
		Entity: parsedEntity,
		Permission: permission,

	}
	expandResponse, err := permissionClient.Expand(context.Background(), expandRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(expandResponse)
}
