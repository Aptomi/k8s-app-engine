package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"fmt"
	"os/exec"
	"aptomi/slinga"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show an object",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var showCmdConfig = &cobra.Command{
	Use:   "config",
	Short: "Show aptomi configuration variables",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		vars := []string{"APTOMI_POLICY", "APTOMI_DB"}

		for _, key := range vars {
			value, _ := os.LookupEnv(key)
			fmt.Println(key + " = " + value)
		}
	},
}

var showCmdPolicy = &cobra.Command{
	Use:   "policy",
	Short: "Show aptomi policy",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var showCmdAllocations = &cobra.Command{
	Use:   "allocations",
	Short: "Show aptomi allocations (what has been allocated and who is using what)",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		usage := slinga.LoadServiceUsageState()
		usage.DrawVisualAndStore()

		command := exec.Command("open", []string{usage.GetVisualFileNamePNG()}...)
		if err := command.Run(); err != nil {
			fmt.Print("Allocations (PNG): " + usage.GetVisualFileNamePNG())
		}
	},
}

func init() {
	showCmd.AddCommand(showCmdConfig)
	showCmd.AddCommand(showCmdPolicy)
	showCmd.AddCommand(showCmdAllocations)

	RootCmd.AddCommand(showCmd)
}
