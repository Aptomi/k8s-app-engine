package revision

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for revision subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revision",
		Short: "revision subcommand",
		Long:  "revision subcommand long",
	}

	cmd.AddCommand(
		newShowCommand(cfg),
	)

	return cmd
}
