package tenancy

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/tui"
	"github.com/Permify/permify-cli/utils"
	v1 "github.com/Permify/permify-go/generated/base/v1"
)

// DeleteCmd - implements tenant delete api
type DeleteCmd struct {
	Command string
}

// Cmd - create command
func (dc *DeleteCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   dc.Command,
		Short: "run delete request",
		Run:  dc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("id", "i", "", "tenant id")
	return cmd
}

func (dc *DeleteCmd) Run(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		newID, err := tui.StringPrompt("Enter id for tenant", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newID== "" {
			log.Error("id must not be empty")
			os.Exit(1)
		}
		id = newID
	}

	tenancyClient := Client()
	deleteRequest := &v1.TenantDeleteRequest{
		Id: id,
	}
	deleteResponse, err := tenancyClient.Delete(context.Background(), deleteRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(deleteResponse)
}
