package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var Version string
var GitCommit string
var BuildTime string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Aptomi version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Aptomi version: %s\n       git commit: %s\n       built: %s\n", Version, GitCommit, BuildTime)
	},
}
