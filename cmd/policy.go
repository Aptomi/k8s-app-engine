package cmd

import (
	"fmt"
	"github.com/Frostman/aptomi/pkg/slinga"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// For apply command
var noop bool
var show bool
var full bool
var verbose bool
var emulateDeployment bool

// For reset command
var force bool

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

		nextUsageState := slinga.NewServiceUsageState(&policy, &dependencies, &users)
		nextUsageState.PrintSummary()
		err := nextUsageState.ResolveAllDependencies()

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

		// Apply changes (if emulateDeployment == true --> we set noop to skip deployment part)
		diff.Apply(noop || emulateDeployment)

		// If everything is successful, then increment revision and save run
		// if emulateDeployment == true --> we set noop to false to write state on disk)
		revision := slinga.GetLastRevision(slinga.GetAptomiBaseDir())
		diff.ProcessSuccessfulExecution(revision, full, noop && !emulateDeployment)
	},
}

var policyCmdReset = &cobra.Command{
	Use:   "reset",
	Short: "Reset policy and delete all objects in it",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if force {
			slinga.ResetAptomiState()
		} else {
			fmt.Println("This will erase everything under " + slinga.GetAptomiBaseDir())
			fmt.Println("No action is taken. If you are sure, use --force to delete all the data")
		}
	},
}

func init() {
	policyCmd.AddCommand(policyCmdApply)
	policyCmd.AddCommand(policyCmdReset)
	RootCmd.AddCommand(policyCmd)

	// Flags for the apply command
	policyCmdApply.Flags().BoolVarP(&noop, "noop", "n", false, "Process a policy, but do no apply changes (noop mode)")
	policyCmdApply.Flags().BoolVarP(&full, "full", "f", false, "Fully re-evaluate the policy. Create new revision. Re-create missing instances (if they were manually deleted from the underlying cloud), update running instances ")
	policyCmdApply.Flags().BoolVarP(&show, "show", "s", false, "Display a picture, showing how policy will be evaluated and applied")
	policyCmdApply.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose information in the output")
	policyCmdApply.Flags().BoolVarP(&emulateDeployment, "emulate", "e", false, "Process a policy, do not deploy anything (emulate deployment), save state to the database")

	// Flags for the reset command
	policyCmdReset.Flags().BoolVarP(&force, "force", "f", false, "Reset policy. Delete all files and don't ask for a confirmation")
}
