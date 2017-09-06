package main

import (
	"fmt"
	. "github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// For apply command
var noop bool
var full bool
var newrevision bool
var verbose bool
var emulateDeployment bool

// For reset command
var force bool

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Process policy and execute an action",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic(err)
		}
	},
}

var policyCmdApply = &cobra.Command{
	Use:   "apply",
	Short: "Process policy and apply changes (supports noop mode)",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the previous usage state (for now it's just empty)
		prevState := resolve.NewPolicyResolution()

		// Generate the next usage state
		policyDir := GetAptomiPolicyDir()
		store := NewFileLoader(policyDir)
		policy := NewPolicyNamespace()

		objects, err := store.LoadObjects()
		if err != nil {
			log.Panicf("Error while loading Policy objects: %v", err)
		}

		for _, object := range objects {
			policy.AddObject(object)
		}

		if err != nil {
			log.Panicf("Cannot load policy from %s with error: %v", policyDir, err)
		}

		externalData := external.NewData(
			users.NewUserLoaderFromLDAP(GetAptomiPolicyDir()),
			secrets.NewSecretLoaderFromDir(GetAptomiPolicyDir()),
		)
		resolver := resolve.NewPolicyResolver(policy, externalData)
		nextState, eventLog, err := resolver.ResolveAllDependencies()
		if err != nil {
			log.Panicf("Cannot resolve policy: %v %v %v", err, nextState, prevState)
		}
		eventLog.Save(&eventlog.HookStdout{})

		fmt.Println("Success")
		fmt.Println("Components:", len(nextState.ComponentInstanceMap))

		// Process differences
		// diff := NewRevisionDiff(nextState, prevState)
		// diff.AlterDifference(full)
		// diff.StoreDiffAsText(verbose)

		// Print on screen
		// fmt.Print(diff.DiffAsText)

		// Generate pictures
		// visual := NewPolicyVisualizationImage(diff)
		// visual.GetImageForRevisionPrev() // just call and don't save
		// visual.GetImageForRevisionNext() // just call and don't save
		// visual.GetImageForRevisionDiff() // just call and don't save

		// Apply changes (if emulateDeployment == true --> we set noop to skip deployment part)

		/*
			apply := NewEngineApply(diff)
			if !(noop || emulateDeployment) {
				err := apply.Apply()
				apply.SaveLog()
				if err != nil {
					log.Panicf("Cannot apply policy: %v", err)
				}
			}
		*/
		// Save new resolved state in the last run directory
		// resolver.SavePolicyResolution()

		// If everything is successful, then increment revision and save run
		// if emulateDeployment == true --> we set noop to false to write state on disk)
		// revision := GetLastRevision(GetAptomiBaseDir())
		// diff.ProcessSuccessfulExecution(revision, newrevision, noop && !emulateDeployment)
		fmt.Println("******************")
		fmt.Println("* The end!       *")
		fmt.Println("* of old CLI     *")
		fmt.Println("******************")
	},
}

var policyCmdReset = &cobra.Command{
	Use:   "reset",
	Short: "Reset policy and delete all objects in it",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if force {
			ResetAptomiState()
		} else {
			fmt.Println("This will erase everything under " + GetAptomiBaseDir())
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
	policyCmdApply.Flags().BoolVarP(&full, "full", "f", false, "Re-create missing instances (if they were manually deleted from the underlying cloud), update running instances")
	policyCmdApply.Flags().BoolVarP(&newrevision, "newrevision", "c", false, "Create new revision, irrespective of whether there are changes in the policy or not")
	policyCmdApply.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose information in the output")
	policyCmdApply.Flags().BoolVarP(&emulateDeployment, "emulate", "e", false, "Process a policy, do not deploy anything (emulate deployment), save state to the database")

	// Flags for the reset command
	policyCmdReset.Flags().BoolVarP(&force, "force", "f", false, "Reset policy. Delete all files and don't ask for a confirmation")
}
