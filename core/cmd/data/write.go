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

// WriteCmd - implements data write api
type WriteCmd struct {
	Command string
}

// Cmd - create command
func (wc *WriteCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   wc.Command,
		Short: "run write request",
		Run:  wc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("entity", "e", "", "entity identifier specified as - <type>:<id>")
	cmd.Flags().StringP("relation", "r", "", "relation between entity and subject")
	cmd.Flags().StringP("subject", "s", "", "subject identifier specified as - <type>:<id>#relation (relation is optional)")
	return cmd
}

func (wc *WriteCmd) Run(cmd *cobra.Command, args []string) {
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
		if newRelation == "" {
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

	schemaVersion, _ := cmd.Flags().GetString("schema")
	dataClient := Client()
	writeRequest := &v1.DataWriteRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.DataWriteRequestMetadata{
			SchemaVersion: schemaVersion,
		},
		Tuples: []*v1.Tuple{{
			Entity: parsedEntity,
			Relation: relation,
			Subject: parsedSubject,
		}},
	}
	writeResponse, err := dataClient.Write(context.Background(), writeRequest)	
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(writeResponse)
}
