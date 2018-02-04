package state

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for state subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "state subcommand",
		Long:  "state subcommand long",
	}

	cmd.AddCommand(
		newResetCommand(cfg),
	)

	return cmd
}
