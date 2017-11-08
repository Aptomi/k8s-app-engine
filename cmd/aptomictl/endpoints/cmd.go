package endpoints

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/spf13/cobra"
)

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
