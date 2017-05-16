package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"aptomi/slinga"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Process policy and execute an action",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var policyCmdApply = &cobra.Command{
	Use:   "apply",
	Short: "Evaluate a policy and apply changes",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		policyDir := slinga.GetAptomiPolicyDir()

		policy := slinga.LoadPolicyFromDir(policyDir)
		users := slinga.LoadUsersFromDir(policyDir)
		dependencies := slinga.LoadDependenciesFromDir(policyDir)

		usageState := slinga.NewServiceUsageState(&policy, &dependencies)
		err := usageState.ResolveUsage(&users)

		if err != nil {
			log.Fatal(err)
		}

		usageState.SaveServiceUsageState()
	},
}

var policyCmdNoop = &cobra.Command{
	Use:   "noop",
	Short: "Evaluate a policy and print expected changes (noop mode)",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	policyCmd.AddCommand(policyCmdApply)
	policyCmd.AddCommand(policyCmdNoop)

	RootCmd.AddCommand(policyCmd)
}
