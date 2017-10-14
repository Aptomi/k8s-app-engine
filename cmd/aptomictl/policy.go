package main

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	policyCmd = &cobra.Command{
		Use:   "policy",
		Short: "policy subcommand",
		Long:  "policy subcommand long",
	}
	policyShowCmd = &cobra.Command{
		Use:   "show",
		Short: "policy show",
		Long:  "policy show long",

		Run: func(cmd *cobra.Command, args []string) {
			err := client.Show(cfg)
			if err != nil {
				panic(fmt.Sprintf("Error while showing policy: %s", err))
			}
		},
	}
	policyApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "apply policy files",
		Long:  "apply policy files long",

		Run: func(cmd *cobra.Command, args []string) {
			err := client.Apply(cfg)
			if err != nil {
				panic(fmt.Sprintf("Error while applying specified policy: %s", err))
			}
		},
	}
)

func init() {
	policyApplyCmd.Flags().StringSliceP("policyPaths", "f", make([]string, 0), "Paths to files, dirs with policy to apply")
	err := viper.BindPFlag("apply.policyPaths", policyApplyCmd.Flags().Lookup("policyPaths"))
	if err != nil {
		panic(err) // todo is it ok to panic here?
	}

	policyCmd.AddCommand(policyApplyCmd, policyShowCmd)
	aptomiCtlCmd.AddCommand(policyCmd)
}
