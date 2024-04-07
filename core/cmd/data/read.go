package data

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

// ReadCmd - implements data read api
type ReadCmd struct {
	Command string
}

// Cmd - check command
func (rc *ReadCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rc.Command,
		Short: "run read request",
		Run:  func(cmd *cobra.Command, args []string) {
			cmd.Help() 
		},
		PreRun: utils.CheckIfUnknownSubcommand,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	entityCommand := ReadRelationsCmd{"relations"}
	subjectCommand := ReadAttributesCmd{"attributes"}
	cmd.AddCommand(entityCommand.Cmd())
	cmd.AddCommand(subjectCommand.Cmd())
	return cmd
}

type ReadRelationsCmd struct {
	Command string
}

func (rr *ReadRelationsCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: rr.Command,
		Short: "run read relationships request",
		Run: rr.Run,
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringP("entity", "e", "", "entity filter specified as - <type>:<id>")
	cmd.Flags().StringP("relation", "r", "", "relation filter")
	cmd.Flags().StringP("subject", "s", "", "subject filter specified as - <type>:<id>#relation (relation is optional)")
	return cmd
}

func (rr *ReadRelationsCmd) Run(cmd *cobra.Command, args []string) {
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

	relation, _ := cmd.Flags().GetString("relation")
	if relation == "" {
		newRelation, err := tui.StringPrompt("Enter relation", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newRelation== "" {
			log.Error("relation must not be empty")
			os.Exit(1)
		}
		relation = newRelation
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

	dataClient := Client()
	relationsRequest := &v1.RelationshipReadRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.RelationshipReadRequestMetadata{},
		Filter: &v1.TupleFilter{
			Entity: &v1.EntityFilter{
				Type: parsedEntity.Type,
				Ids: []string{parsedEntity.Id},
			},
			Relation: relation,
			Subject: &v1.SubjectFilter{
				Type: parsedSubject.Type,
				Ids: []string{parsedSubject.Id},
			},
		},
	}
	relationResponse, err := dataClient.ReadRelationships(context.Background(), relationsRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(relationResponse)
}

type ReadAttributesCmd struct {
	Command string
}

func (ra *ReadAttributesCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: ra.Command,
		Short: "run lookup subject request",
		Run: ra.Run,
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringP("entity", "e", "", "entity filter specified as - <type>:<id>")
	cmd.Flags().StringP("relation", "r", "", "relation filter")
	return cmd
}

func (ra *ReadAttributesCmd) Run(cmd *cobra.Command, args []string) {
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

	attribute, _ := cmd.Flags().GetString("attribute")
	if attribute == "" {
		newAttribute, err := tui.StringPrompt("Enter attribute", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newAttribute == "" {
			log.Error("attribute must not be empty")
			os.Exit(1)
		}
		attribute = newAttribute
	}

	dataClient := Client()
	attributeRequest := &v1.AttributeReadRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.AttributeReadRequestMetadata{},
		Filter: &v1.AttributeFilter{
			Entity: &v1.EntityFilter{
				Type: parsedEntity.Type,
				Ids: []string{parsedEntity.Id},
			},
			Attributes: []string{attribute},
		},
	}
	attributeResponse, err := dataClient.ReadAttributes(context.Background(), attributeRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(attributeResponse)
}