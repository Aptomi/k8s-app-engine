package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"aptomi/slinga"
)

var noop bool

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
	Short: "Process policy and apply changes (supports noop mode)",
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

		// TODO: implement
		if noop {
			// do not apply changes
		} else {
			// apply changes
		}

		usageState.SaveServiceUsageState()
	},
}

func init() {
	policyCmd.AddCommand(policyCmdApply)
	RootCmd.AddCommand(policyCmd)

	policyCmdApply.Flags().BoolVarP(&noop, "noop", "n", false, "Process a policy, but do no apply changes (noop mode)")
}
