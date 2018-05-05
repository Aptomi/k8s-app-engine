package server

import (
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/server"
	"github.com/spf13/cobra"
)

// NewServerCommand returns instance of cobra command that starts the server
func NewServerCommand(cfg *config.Server) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start Aptomi server",
		Long:  "Start Aptomi server",

		Run: func(cmd *cobra.Command, args []string) {
			server.NewServer(cfg).Start()
		},
	}
}
