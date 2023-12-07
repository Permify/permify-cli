package flags

import (
	"github.com/spf13/cobra"
)

// RegisterRootFlags - Define and registers permify CLI flags
func RegisterRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("endpoint", "", "permify gRPC endpoint")
}
