package cmd

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	"github.com/spf13/cobra"
	log "github.com/Sirupsen/logrus"
)

var noop bool
var show bool
var verbose bool

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

		nextUsageState := slinga.NewServiceUsageState(&policy, &dependencies)
		err := nextUsageState.ResolveUsage(&users)

		if err != nil {
			log.Panicf("Cannot resolve usage: %v", err)
		}

		// Process differences
		diff := nextUsageState.CalculateDifference(&prevUsageState)

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
	policyCmdApply.Flags().BoolVarP(&show, "show", "s", false, "Display a picture, showing how policy will be evaluated and applied")
	policyCmdApply.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose information in the output")
}
