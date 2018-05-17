package dependency

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for dependency subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dependency",
		Short: "Dependency subcommand",
		Long:  "Dependency subcommand long",
	}

	cmd.AddCommand(
		newStatusCommand(cfg),
	)

	return cmd
}
