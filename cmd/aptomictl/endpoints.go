package main

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/client"
	"github.com/spf13/cobra"
)

var (
	endpointsCmd = &cobra.Command{
		Use:   "endpoints",
		Short: "endpoints subcommand",
		Long:  "endpoints subcommand long",
	}
	endpointsShowCmd = &cobra.Command{
		Use:   "show",
		Short: "endpoints show",
		Long:  "endpoints show long",

		Run: func(cmd *cobra.Command, args []string) {
			err := client.Endpoints(cfg)
			if err != nil {
				panic(fmt.Sprintf("Error while showing endpoints: %s", err))
			}
		},
	}
)

func init() {
	endpointsCmd.AddCommand(endpointsShowCmd)
	aptomiCtlCmd.AddCommand(endpointsCmd)
}
