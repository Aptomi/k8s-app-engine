package policy

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for policy subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "policy subcommand",
		Long:  "policy subcommand long",
	}

	cmd.AddCommand(
		newShowCommand(cfg),
		newApplyCommand(cfg),
		newDeleteCommand(cfg),
	)

	return cmd
}
