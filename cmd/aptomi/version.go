package main

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/version"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Aptomi version",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()
		fmt.Printf("Aptomi version: %s\n       git commit: %s\n       built: %s\n", info.GitVersion, info.GitCommit, info.BuildDate)
	},
}
