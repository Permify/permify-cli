package permission

import (
	"context"
	"os"

	"github.com/charmbracelet/log"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/core/config"
	"github.com/Permify/permify-cli/tui"
	"github.com/Permify/permify-cli/utils"
)

// CheckCmd - implements permission check api
type CheckCmd struct {
	Command string
}

// Cmd - check command
func (cc *CheckCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cc.Command,
		Short: "run check request",
		Run:  cc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("entity", "e", "", "entity identifier specified as - <type>:<id>")
	cmd.Flags().StringP("permission", "p", "", "permission to check")
	cmd.Flags().StringP("subject", "s", "", "subject identifier specified as - <type>:<id>#relation (relation is optional)")
	cmd.Flags().Int32("depth", 50, "depth of the check must be >= 3. Default: 50")
	return cmd
}

func (cc *CheckCmd) Run(cmd *cobra.Command, args []string) {
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
	if permission == "" {
		newPermission, err := tui.StringPrompt("Enter permission to check", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newPermission == "" {
			log.Error("permission must not be empty")
			os.Exit(1)
		}
		permission = newPermission
	}

	subject, _ := cmd.Flags().GetString("subject")
	if subject == "" {
		newSubject, err := tui.StringPrompt("Enter subject string (relation is optional)", "<type>:<id>#<relation>", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		subject = newSubject 
	}
	parsedSubject, err := utils.ParseSubject(subject)

	schemaVersion, _ := cmd.Flags().GetString("schema")
	depth, _ := cmd.Flags().GetInt32("depth")

	permissionClient := Client()
	checkRequest := &v1.PermissionCheckRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.PermissionCheckRequestMetadata{
			SchemaVersion: schemaVersion,
			Depth: depth,
		},
		Entity: parsedEntity,
		Permission: permission,
		Subject: parsedSubject,
	}
	checkResponse, err := permissionClient.Check(context.Background(), checkRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(checkResponse)
}
