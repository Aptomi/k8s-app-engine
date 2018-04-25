package gen

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for cluster subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "Gen subcommand",
		Long:  "Gen subcommand long",
	}

	cmd.AddCommand(
		newClusterCommand(cfg),
		newUserCommand(cfg),
	)

	return cmd
}
