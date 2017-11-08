package policy

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "policy subcommand",
		Long:  "policy subcommand long",
	}

	cmd.AddCommand(
		newShowCommand(cfg),
		newApplyCommand(cfg),
	)

	return cmd
}
