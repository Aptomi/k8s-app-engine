package policy

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for policy subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Policy subcommand",
		Long:  "Policy subcommand long",
	}

	cmd.AddCommand(
		newShowCommand(cfg),                       // show
		newHandlePolicyChangesCommand(cfg, true),  // apply
		newHandlePolicyChangesCommand(cfg, false), // delete
	)

	return cmd
}
