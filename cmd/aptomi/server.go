package main

import (
	"github.com/Aptomi/aptomi/pkg/slinga"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		slinga.Serve()
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
