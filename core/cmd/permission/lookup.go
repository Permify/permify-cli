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

// LookupCmd - implements permission check api
type LookupCmd struct {
	Command string
}

// Cmd - check command
func (lc *LookupCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   lc.Command,
		Short: "run lookup request",
		Run:  func(cmd *cobra.Command, args []string) {
			cmd.Help() 
		},
		PreRun: utils.CheckIfUnknownSubcommand,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	entityCommand := LookupEntityCmd{"entity"}
	subjectCommand := LookupSubjectCmd{"subject"}
	cmd.AddCommand(entityCommand.Cmd())
	cmd.AddCommand(subjectCommand.Cmd())
	return cmd
}

type LookupEntityCmd struct {
	Command string
}

func (le *LookupEntityCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: le.Command,
		Short: "run lookup entity request",
		Run: le.Run,
		Args: cobra.NoArgs,
	}
	cmd.Flags().Int32("depth", 50, "depth of the check must be >= 3")
	cmd.Flags().StringP("type", "t", "", "entity type to lookup")
	cmd.Flags().StringP("permission", "p", "", "permission to check")
	cmd.Flags().StringP("subject", "s", "", "subject identifier specified as - <type>:<id>#relation (relation is optional)")
	return cmd
}

func (le *LookupEntityCmd) Run(cmd *cobra.Command, args []string) {
	schemaVersion, _ := cmd.Flags().GetString("schema")
	depth, _ := cmd.Flags().GetInt32("depth")

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

	entityType, _ := cmd.Flags().GetString("type")
	if permission == "" {
		newEntityType, err := tui.StringPrompt("Enter entity type to lookup", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newEntityType == "" {
			log.Error("entity type must not be empty")
			os.Exit(1)
		}
		entityType = newEntityType
	}

	permissionClient := Client()
	lookupRequest := &v1.PermissionLookupEntityRequest{
		TenantId: config.CliConfig.Tenant,	
		Metadata: &v1.PermissionLookupEntityRequestMetadata{
			SchemaVersion: schemaVersion,
			Depth: depth,
		},
		Subject: parsedSubject,
		EntityType: entityType,
		Permission: permission,
	}
	lookupResponse, err := permissionClient.LookupEntity(context.Background(), lookupRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(lookupResponse)	
}

type LookupSubjectCmd struct {
	Command string
}

func (ls *LookupSubjectCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: ls.Command,
		Short: "run lookup subject request",
		Run: ls.Run,
		Args: cobra.NoArgs,
	}
	cmd.Flags().Int32("depth", 50, "depth of the check must be >= 3")
	cmd.Flags().StringP("permission", "p", "", "permission to check")
	cmd.Flags().StringP("type", "t", "", "subject type to lookup")
	cmd.Flags().StringP("relation", "r", "", "[Optional] subject relation to lookup")
	cmd.Flags().StringP("entity", "e", "", "entity identifier specified as - <type>:<id>")
	return cmd
}

func (ls *LookupSubjectCmd) Run(cmd *cobra.Command, args []string) {
	schemaVersion, _ := cmd.Flags().GetString("schema")
	depth, _ := cmd.Flags().GetInt32("depth")

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

	subjectType, _ := cmd.Flags().GetString("type")
	if permission == "" {
		newSubjectType, err := tui.StringPrompt("Enter subject type to lookup", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newSubjectType == "" {
			log.Error("entity type must not be empty")
			os.Exit(1)
		}
		subjectType = newSubjectType
	}

	subjectRelation, _ := cmd.Flags().GetString("relation")

	permissionClient := Client()
	lookupRequest := &v1.PermissionLookupSubjectRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.PermissionLookupSubjectRequestMetadata{
			SchemaVersion: schemaVersion,
			Depth: depth,
		},
		Entity: parsedEntity,
		Permission: permission,
		SubjectReference: &v1.RelationReference{
			Type: subjectType,
			Relation: subjectRelation,
		},
	}
	lookupResponse, err := permissionClient.LookupSubject(context.Background(), lookupRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(lookupResponse)
}