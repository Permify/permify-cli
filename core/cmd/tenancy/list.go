package tenancy

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/utils"
	v1 "github.com/Permify/permify-go/generated/base/v1"
)

// ListCmd - implements tenant list api
type ListCmd struct {
	Command string
}

// Cmd - create command
func (lc *ListCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   lc.Command,
		Short: "run list request",
		Run:  lc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	return cmd
}

func (lc *ListCmd) Run(cmd *cobra.Command, args []string) {
	tenancyClient := Client()
	listRequest := &v1.TenantListRequest{}
	listResponse, err := tenancyClient.List(context.Background(), listRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(listResponse)
}
