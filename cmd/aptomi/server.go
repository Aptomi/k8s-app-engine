package main

import (
	"github.com/Aptomi/aptomi/pkg/server"
	"github.com/spf13/cobra"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "start Aptomi server",
		Long:  "start Aptomi server",

		Run: func(cmd *cobra.Command, args []string) {
			server.NewServer(cfg).Start()
		},
	}
)

func init() {
	aptomiCmd.AddCommand(serverCmd)
}
