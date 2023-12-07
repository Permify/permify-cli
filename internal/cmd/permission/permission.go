package permission

import (
	"github.com/spf13/cobra"
)

// NewPermissionCommand - Creates new permission command
func NewPermissionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permission",
		Short: "",
	}
	return cmd
}
