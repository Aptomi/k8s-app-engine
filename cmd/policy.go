package cmd

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var noop bool
var show bool
var full bool
var verbose bool
var trace bool

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Process policy and execute an action",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var policyCmdApply = &cobra.Command{
	Use:   "apply",
	Short: "Process policy and apply changes (supports noop mode)",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the previous usage state
		prevUsageState := slinga.LoadServiceUsageState()

		// Generate the next usage state
		policyDir := slinga.GetAptomiPolicyDir()

		policy := slinga.LoadPolicyFromDir(policyDir)
		users := slinga.LoadUsersFromDir(policyDir)
		dependencies := slinga.LoadDependenciesFromDir(policyDir)
		dependencies.SetTrace(trace)

		nextUsageState := slinga.NewServiceUsageState(&policy, &dependencies, &users)
		err := nextUsageState.ResolveUsage()

		if err != nil {
			log.Panicf("Cannot resolve usage: %v", err)
		}

		// Process differences
		diff := nextUsageState.CalculateDifference(&prevUsageState)
		diff.AlterDifference(full)

		// Print on screen
		diff.Print(verbose)

		// Generate pictures, if needed
		if show {
			visual := slinga.NewPolicyVisualization(diff)
			visual.DrawAndStore()
			visual.OpenInPreview()
		}

		// Apply changes
		diff.Apply(noop)
	},
}

func init() {
	policyCmd.AddCommand(policyCmdApply)
	RootCmd.AddCommand(policyCmd)

	policyCmdApply.Flags().BoolVarP(&noop, "noop", "n", false, "Process a policy, but do no apply changes (noop mode)")
	policyCmdApply.Flags().BoolVarP(&full, "full", "f", false, "In addition to applying changes, re-create missing instances (if they were manually deleted from the underlying cloud) and update running instances")
	policyCmdApply.Flags().BoolVarP(&show, "show", "s", false, "Display a picture, showing how policy will be evaluated and applied")
	policyCmdApply.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose information in the output")
	policyCmdApply.Flags().BoolVarP(&trace, "trace", "t", false, "Trace all dependencies and print how rules got evaluated")
}
