package cmd

import (
	"github.com/spf13/cobra"
	"github.com/golang/glog"
	"github.com/Frostman/aptomi/pkg/slinga"
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
		// Load the previous usage state
		prevUsageState := slinga.LoadServiceUsageState()

		// Generate the next usage state
		policyDir := slinga.GetAptomiPolicyDir()

		policy := slinga.LoadPolicyFromDir(policyDir)
		users := slinga.LoadUsersFromDir(policyDir)
		dependencies := slinga.LoadDependenciesFromDir(policyDir)

		nextUsageState := slinga.NewServiceUsageState(&policy, &dependencies)
		err := nextUsageState.ResolveUsage(&users)

		if err != nil {
			glog.Fatal(err)
		}

		// Process differences
		diff := nextUsageState.ProcessDifference(&prevUsageState)

		if noop {
			// do not apply changes
			diff.Print()
		} else {
			// apply changes
			// TODO: implement

			// save new state
			nextUsageState.SaveServiceUsageState()
		}
	},
}

func init() {
	policyCmd.AddCommand(policyCmdApply)
	RootCmd.AddCommand(policyCmd)

	policyCmdApply.Flags().BoolVarP(&noop, "noop", "n", false, "Process a policy, but do no apply changes (noop mode)")
}
