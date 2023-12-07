package flags

import (
	"github.com/spf13/cobra"
)

// RegisterCheckFlags - Define and registers permify CLI flags
func RegisterCheckFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("context", "", "check request context")
}
