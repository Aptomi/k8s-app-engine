package cmd

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// For apply command
var noop bool
var show bool
var full bool
var verbose bool
var trace bool

// For reset command
// var force bool

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
		baseDir := slinga.GetAptomiBaseDir() + "/aptomi-demo"
		policy := slinga.LoadPolicyFromDir(baseDir)
		users := slinga.LoadUsersFromDir(baseDir)
		dependencies := slinga.LoadDependenciesFromDir(baseDir)
		dependencies.SetTrace(trace)

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

		// Apply changes
		diff.Apply(noop)
	},
}

/*
var policyCmdAdd = &cobra.Command{
	Use:   "add",
	Short: "Add objects to the policy (or edit existing objects)",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var policyCmdDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete objects from the policy",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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
			fmt.Println("Sorry. I won't be doing anything without --force.")
		}
	},
}
*/

func init() {
	policyCmd.AddCommand(policyCmdApply)

	/*
		policyCmd.AddCommand(policyCmdAdd)
		policyCmd.AddCommand(policyCmdDelete)
		policyCmd.AddCommand(policyCmdReset)
	*/

	/*
		for k := range slinga.AptomiObjectsCanBeModified {
			command := &cobra.Command{
				Use:   k,
				Short: fmt.Sprintf("Add one or more %s to the policy", k),
				Long:  "",
				Run: func(cmd *cobra.Command, args []string) {
					slinga.AddObjectsToPolicy(slinga.AptomiObjectsCanBeModified[cmd.Use], args...)
				},
			}
			policyCmdAdd.AddCommand(command)
		}

		for k := range slinga.AptomiObjectsCanBeModified {
			command := &cobra.Command{
				Use:   k,
				Short: fmt.Sprintf("Delete one or more %s from the policy", k),
				Long:  "",
				Run: func(cmd *cobra.Command, args []string) {
					slinga.RemoveObjectsFromPolicy(slinga.AptomiObjectsCanBeModified[cmd.Use], args...)
				},
			}
			policyCmdDelete.AddCommand(command)
		}
	*/

	RootCmd.AddCommand(policyCmd)

	// Flags for the apply command
	policyCmdApply.Flags().BoolVarP(&noop, "noop", "n", false, "Process a policy, but do no apply changes (noop mode)")
	policyCmdApply.Flags().BoolVarP(&full, "full", "f", false, "In addition to applying changes, re-create missing instances (if they were manually deleted from the underlying cloud) and update running instances")
	policyCmdApply.Flags().BoolVarP(&show, "show", "s", false, "Display a picture, showing how policy will be evaluated and applied")
	policyCmdApply.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose information in the output")
	policyCmdApply.Flags().BoolVarP(&trace, "trace", "t", false, "Trace all dependencies and print how rules got evaluated")

	// Flags for the reset command
	// policyCmdReset.Flags().BoolVarP(&force, "force", "f", false, "Reset policy. Delete all files and don't ask for a confirmation")
}
