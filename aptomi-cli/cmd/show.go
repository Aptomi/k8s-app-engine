package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"fmt"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an object",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var showCmdA = &cobra.Command{
	Use:   "vars",
	Short: "Show aptomi variables",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		vars := []string{"APTOMI_POLICY", "APTOMI_DB"}

		for _, key := range vars {
			value, _ := os.LookupEnv(key)
			fmt.Println(key + " = " + value)
		}
	},
}

var showCmdB = &cobra.Command{
	Use:   "policy",
	Short: "Show aptomi policy",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var showCmdC = &cobra.Command{
	Use:   "allocations",
	Short: "Show aptomi allocations (what has been allocated and who is using what)",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	showCmd.AddCommand(showCmdA)
	showCmd.AddCommand(showCmdB)
	showCmd.AddCommand(showCmdC)

	RootCmd.AddCommand(showCmd)
}
