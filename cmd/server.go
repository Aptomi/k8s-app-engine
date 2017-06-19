package cmd

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		slinga.Serve("", 8080)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
}
