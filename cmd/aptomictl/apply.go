package main

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/slinga/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply policy files",
		Long:  "",

		Run: func(cmd *cobra.Command, args []string) {
			err := client.Apply(viper.GetViper())
			if err != nil {
				panic(fmt.Sprintf("Error while applying specified policy: %s", err))
			}
		},
	}
)

func init() {
	applyCmd.Flags().StringSliceP("policyPaths", "f", make([]string, 0), "Paths to files, dirs with policy to apply")
	viper.BindPFlag("apply.policyPaths", applyCmd.Flags().Lookup("policyPaths"))

	AptomiCtlCmd.AddCommand(applyCmd)
}
