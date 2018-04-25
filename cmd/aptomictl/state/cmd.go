package state

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for state subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "State subcommand",
		Long:  "State subcommand long",
	}

	cmd.AddCommand(
		newEnforceCommand(cfg),
	)

	return cmd
}
