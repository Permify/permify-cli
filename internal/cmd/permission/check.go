package permission

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/spf13/cobra"

	"github.com/Permify/permify-cli/internal/client"
	"github.com/Permify/permify-cli/internal/cmd/flags"
	"github.com/Permify/permify-cli/pkg/tuple"
)

// NewCheckCommand - Creates new check command
func NewCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <tenant> <entity:id> <permission> <subject:id>",
		Short: "",
		RunE:  check(),
		Args:  cobra.ExactArgs(4),
	}

	flags.RegisterCheckFlags(cmd)

	return cmd
}

func check() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, err := client.New("localhost:3478")
		if err != nil {
			return err
		}

		entity, err := tuple.E(args[1])
		if err != nil {
			return err
		}

		ear, err := tuple.EAR(args[3])
		subject := &v1.Subject{
			Type:     ear.GetEntity().GetType(),
			Id:       ear.GetEntity().GetId(),
			Relation: ear.GetRelation(),
		}

		response, err := c.Permission.Check(context.Background(), &v1.PermissionCheckRequest{
			TenantId: args[0],
			Metadata: &v1.PermissionCheckRequestMetadata{
				SchemaVersion: "",
				SnapToken:     "",
				Depth:         5,
			},
			Entity:     entity,
			Permission: args[2],
			Subject:    subject,
		})
		if err != nil {
			return err
		}

		status := response.GetCan()
		checkCount := response.GetMetadata().GetCheckCount()

		table := tablewriter.NewWriter(os.Stdout)

		// Displaying the status
		switch status {
		case v1.CheckResult_CHECK_RESULT_ALLOWED:
			table.Append([]string{"Result", "ALLOWED"})
		case v1.CheckResult_CHECK_RESULT_DENIED:
			table.Append([]string{"Result", "DENIED"})
		default:
			table.Append([]string{"Result", "UNKNOWN"})
		}

		// Displaying the check count
		table.Append([]string{"Check Count", fmt.Sprintf("%d", checkCount)})

		table.Render() // Send output

		return nil
	}
}
