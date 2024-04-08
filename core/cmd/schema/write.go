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

// WriteCmd - implements schema read api
type WriteCmd struct {
	Command string
}

// Cmd - read command
func (wc *WriteCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   wc.Command,
		Short: "run read request",
		Run:  wc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("file", "f", "", "perm schema file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func (wc *WriteCmd) Run(cmd *cobra.Command, args []string) {
	file, _ := cmd.Flags().GetString("file")
	schema, err := utils.ReadFileToString(file)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	schemaClient := Client()
	writeRequest := &v1.SchemaWriteRequest{
		TenantId: config.CliConfig.Tenant,
		Schema: schema,
	}
	writeResponse, err := schemaClient.Write(context.Background(), writeRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(writeResponse)
}
