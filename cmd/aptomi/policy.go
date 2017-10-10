package main

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

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
		policyDir := db.GetAptomiPolicyDir()
		store := lang.NewFileLoader(policyDir)
		policy := lang.NewPolicy()

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
			users.NewUserLoaderFromLDAP(db.GetAptomiPolicyDir()),
			secrets.NewSecretLoaderFromDir(db.GetAptomiPolicyDir()),
		)
		resolver := resolve.NewPolicyResolver(policy, externalData)
		resolution, eventLog, err := resolver.ResolveAllDependencies()
		if err != nil {
			eventLog.Save(&event.HookStdout{})
			log.Panicf("Cannot resolve policy: %v %v %v", err, resolution, prevState)
		}
		eventLog.Save(&event.HookStdout{})

		fmt.Println("Success")
		fmt.Println("Components:", len(resolution.ComponentInstanceMap))

		// image, _ := visualization.CreateImage(visualization.NewDiagram(policy, resolution, externalData))
		// visualization.OpenImage(image)

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

func init() {
	policyCmd.AddCommand(policyCmdApply)
	RootCmd.AddCommand(policyCmd)
}
