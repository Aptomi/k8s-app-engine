package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an object",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var showCmdConfig = &cobra.Command{
	Use:   "config",
	Short: "Show aptomi configuration variables",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		vars := []string{"APTOMI_DB"}

		for _, key := range vars {
			value, _ := os.LookupEnv(key)
			fmt.Println(key + " = " + value)
		}
	},
}

var showCmdPolicy = &cobra.Command{
	Use:   "policy",
	Short: "Show aptomi policy",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: implement show policy
		fmt.Println("Not implemented")
	},
}

func init() {
	showCmd.AddCommand(showCmdConfig)
	showCmd.AddCommand(showCmdPolicy)

	RootCmd.AddCommand(showCmd)
}
