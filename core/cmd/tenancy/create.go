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

// CreateCmd - implements tenant create api
type CreateCmd struct {
	Command string
}

// Cmd - create command
func (cc *CreateCmd) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cc.Command,
		Short: "run create request",
		Run:  cc.Run,
		Args:  cobra.NoArgs,
	}
	cmd.SetHelpFunc(utils.CmdHelp)
	cmd.Flags().StringP("id", "i", "", "tenant id")
	cmd.Flags().StringP("name", "n", "", "tenant name")
	return cmd
}

func (cc *CreateCmd) Run(cmd *cobra.Command, args []string) {
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

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		newName, err := tui.StringPrompt("Enter name for tenant", "", "")
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		if newName== "" {
			log.Error("name must not be empty")
			os.Exit(1)
		}
		name = newName
	}

	tenancyClient := Client()
	createRequest := &v1.TenantCreateRequest{
		Id: id,
		Name: name,
	}
	createResponse, err := tenancyClient.Create(context.Background(), createRequest)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	utils.PrettyPrint(createResponse)
}
