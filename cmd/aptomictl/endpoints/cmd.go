package endpoints

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand returns cobra command for endpoints subcommand
func NewCommand(cfg *config.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoints",
		Short: "endpoints subcommand",
		Long:  "endpoints subcommand long",
	}

	cmd.AddCommand(
		newShowCommand(cfg),
	)

	return cmd
}
