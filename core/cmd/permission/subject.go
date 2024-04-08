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

// SubjectCmd - implements permission check api
type SubjectCmd struct {
	Command string
}

// Cmd - check command
func (sc *SubjectCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   sc.Command,
		Short: "run lookup request",
		Run:  sc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("entity", "e", "", "entity identifier specified as - <type>:<id>")
	cmd.Flags().StringP("subject", "s", "", "subject identifier specified as - <type>:<id>#relation (relation is optional)")
	cmd.Flags().Int32("depth", 50, "depth of the check must be >= 3")
	cmd.Flags().BoolP("only-permission", "p", false, "return only permissions. Default: false")
	return cmd
}

func (sc *SubjectCmd) Run(cmd *cobra.Command, args []string) {
	schemaVersion, _ := cmd.Flags().GetString("schema")
	depth, _ := cmd.Flags().GetInt32("depth")
	onlyPermission, _ := cmd.Flags().GetBool("only-permission")

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

	permissionClient := Client()
	subjectRequest := &v1.PermissionSubjectPermissionRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.PermissionSubjectPermissionRequestMetadata{
			SchemaVersion: schemaVersion,
			OnlyPermission: onlyPermission,
			Depth: depth,
		},	
		Entity: parsedEntity,
		Subject: parsedSubject,
	}
	subjectResponse, err := permissionClient.SubjectPermission(context.Background(), subjectRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(subjectResponse)
}
