package claim

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for claim subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim subcommand",
		Long:  "Claim subcommand long",
	}

	cmd.AddCommand(
		newStatusCommand(cfg),
		newEndpointsCommand(cfg),
	)

	return cmd
}
