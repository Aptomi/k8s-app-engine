package main

import (
	"github.com/Aptomi/aptomi/pkg/slinga/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		server.NewServer(viper.GetViper()).Start()
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
