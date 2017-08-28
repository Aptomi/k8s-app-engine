package main

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/slinga/client"
	"github.com/spf13/cobra"
)

var (
	policyPaths []string

	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply policy files",
		Long:  "",

		Run: func(cmd *cobra.Command, args []string) {
			err := client.Apply(policyPaths)
			if err != nil {
				panic(fmt.Sprintf("Error while applying specified policy: %s", err))
			}
		},
	}
)

func init() {
	applyCmd.Flags().StringSliceVarP(&policyPaths, "policyPaths", "f", make([]string, 0), "Paths to files, dirs with policy to apply")

	AptomiCtlCmd.AddCommand(applyCmd)
}
