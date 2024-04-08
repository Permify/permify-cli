package schema

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/core/config"
	"github.com/Permify/permify-cli/utils"
	v1 "github.com/Permify/permify-go/generated/base/v1"
)

// ReadCmd - implements schema read api
type ReadCmd struct {
	Command string
}

// Cmd - read command
func (rc *ReadCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rc.Command,
		Short: "run read request",
		Run:  rc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	return cmd
}

func (rc *ReadCmd) Run(cmd *cobra.Command, args []string) {
	schemaVersion, _ := cmd.Flags().GetString("schema")
	schemaClient := Client()
	readRequest := &v1.SchemaReadRequest{
		TenantId: config.CliConfig.Tenant,
		Metadata: &v1.SchemaReadRequestMetadata{
			SchemaVersion: schemaVersion,
		},
	}
	readResponse, err := schemaClient.Read(context.Background(), readRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(readResponse)
}
