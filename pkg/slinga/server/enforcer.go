package server

import (
	"github.com/Aptomi/aptomi/pkg/slinga/db"
	//"github.com/Aptomi/aptomi/pkg/slinga/engine/apply"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/server/store"
	//"github.com/Aptomi/aptomi/pkg/slinga/visualization"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/plugin/helm"
	log "github.com/Sirupsen/logrus"
	"runtime/debug"
	"time"
)

type Enforcer struct {
	store        store.ServerStore
	externalData *external.Data
}

func NewEnforcer(store store.ServerStore, data *external.Data) *Enforcer {
	return &Enforcer{store, data}
}

func logError(err interface{}) {
	log.Errorf("Error while enforcing policy: %s", err)

	// todo make configurable
	debug.PrintStack()
}

func (e *Enforcer) Enforce() error {
	for {
		err := e.enforce()
		if err != nil {
			logError(err)
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (e *Enforcer) enforce() error {
	defer func() {
		if err := recover(); err != nil {
			logError(err)
		}
	}()

	desiredPolicy, desiredPolicyGen, err := e.store.GetPolicy(object.LastGen)
	if err != nil {
		return fmt.Errorf("Error while getting desiredPolicy: %s", err)
	}

	// skip policy enforcement if no policy found
	if desiredPolicy == nil {
		//todo log
		return nil
	}

	actualState, err := e.store.GetActualState()
	if err != nil {
		return fmt.Errorf("Error while getting actual state: %s", err)
	}

	resolver := resolve.NewPolicyResolver(desiredPolicy, externalData)
	desiredState, eventLog, err := resolver.ResolveAllDependencies()
	if err != nil {
		return fmt.Errorf("Cannot resolve desiredPolicy: %v %v %v", err, desiredState, actualState)
	}

	eventLog.Save(&event.HookStdout{})

	nextRevision, err := e.store.NextRevision(desiredPolicyGen)
	if err != nil {
		return fmt.Errorf("Unable to get next revision: %s", err)
	}

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState, nextRevision.GetGeneration())

	// todo add check that policy gen not changed (always create new revision if policy gen changed)
	if !stateDiff.IsChanged() {
		// todo
		log.Infof("No changes")
		return nil
	}
	//todo
	log.Infof("Changes")
	// todo if policy gen changed, we still need to save revision but with progress == done

	//todo remove debug log
	for _, action := range stateDiff.Actions {
		fmt.Println(action)
	}

	// Save revision
	err = e.store.SaveRevision(nextRevision)
	if err != nil {
		return fmt.Errorf("Error while saving new revision: %s", err)
	}

	// todo generate diagrams
	//prefDiagram := visualization.NewDiagram(actualPolicy, actualState, externalData)
	//newDiagram := visualization.NewDiagram(desiredPolicy, desiredState, externalData)
	//deltaDiagram := visualization.NewDiagramDelta(desiredPolicy, desiredState, actualPolicy, actualState, externalData)
	//visualization.CreateImage(...) for all diagrams

	// Build plugins
	helmIstio := helm.NewPlugin()
	plugins := plugin.NewRegistry(
		[]plugin.DeployPlugin{helmIstio},
		[]plugin.ClustersPostProcessPlugin{helmIstio},
	)

	actualPolicy, err := e.getActualPolicy()
	if err != nil {
		return fmt.Errorf("Error while getting actual policy: %s", err)
	}

	applier := apply.NewEngineApply(desiredPolicy, desiredState, actualPolicy, actualState, e.store.ActualStateUpdater(), externalData, plugins, stateDiff.Actions)
	resolution, eventLog, err := applier.Apply()

	eventLog.Save(&event.HookStdout{})

	if err != nil {
		return fmt.Errorf("Error while applying new revision: %s", err)
	}
	log.Infof("Applied new revision with resolution: %v", resolution)

	return nil
}

func (e *Enforcer) getActualPolicy() (*lang.Policy, error) {
	currRevision, err := e.store.GetRevision(object.LastGen)
	if err != nil {
		return nil, fmt.Errorf("Unable to get current revision: %s", err)
	}

	// it's just a first revision
	if currRevision == nil {
		return lang.NewPolicy(), nil
	}

	actualPolicy, _, err := e.store.GetPolicy(currRevision.Policy)
	if err != nil {
		return nil, fmt.Errorf("Unable to get actual policy: %s", err)
	}

	return actualPolicy, nil
}

/*
func (ctl *RevisionControllerImpl) CheckState() error {
	policy, err := ctl.policyCtl.GetPolicy(object.LastGen)

	// Background applier
	// [1] Calculate current desired state (run resolver // resolver.ResolveAllDependencies())
	// 	 1. Use latest policy version and latest external data
	// [2] Load current actual state
	// [3] Compare actual and desired state, calculate diff and actions
	//   1. Note(!): actual state will not have an associated policy
	//               do not use Prev.Policy in diff or apply
	// [4] If hasChanges =>
	// 				new Revision,
	//				attach resolution event log to the revision (may be attach to RevisionSummary)
	// 		else return no new revision needed
	// [5] Applies executes all actions and updates actual state, if/as needed
	//   1. Once action has been executed, save its status and event log to DB
	// [6] Mark revision as "OK" if all actions were completed without errors
	// [7] Keep RevisionSummary
	// 	 1. process text policy diff, add to RevisionSummary
	// 	 1. generate new charts - NewPolicyVisualization, add to RevisionSummary
	// 	 1. save text diff for component instances into RevisionSummary

	// Load the previous usage state (for now it's just empty), it's for now ActualState
	prevState := resolve.NewPolicyResolution()

	externalData := external.NewData(
		users.NewUserLoaderFromLDAP(db.GetAptomiPolicyDir()),
		secrets.NewSecretLoaderFromDir(db.GetAptomiPolicyDir()),
	)
	resolver := resolve.NewPolicyResolver(policy, externalData)
	resolution, eventLog, err := resolver.ResolveAllDependencies()
	if err != nil {
		eventLog.Save(&event.HookStdout{})
		return fmt.Errorf("Cannot resolve policy: %v %v %v", err, resolution, prevState)
	}
	eventLog.Save(&event.HookStdout{})

	fmt.Println("Success")
	fmt.Println("Components:", len(resolution.ComponentInstanceMap))

	return nil
}

/*
	PolicyResolver should return PolicyResolution
	Remove revision from the engine
	Revision in controller package (created by Sergey) = policyResolution + actions
*/

/*
			// Get loader for external users
			userLoader := NewAptomiUserLoader()

			// Load the previous usage state
			prevState := resolve.LoadRevision()

			policy := ... NewPolicy()

			resolver := resolve.NewPolicyResolver(policy, userLoader)
			nextState, err := resolver.ResolveAllDependencies()
			if err != nil {
				return fmt.Errorf("Cannot resolve policy: %v", err)
			}

			// Process differences
			diff := NewRevisionDiff(nextState, prevState)
			diff.AlterDifference(full)
			diff.StoreDiffAsText(verbose)

			// Print on screen
			fmt.Print(diff.DiffAsText)

			// Generate pictures (see new API :)

			// Save new resolved state in the last run directory
			resolver.SavePolicyResolution() -> this called before:

					revision := NewRevision(resolver.policy, resolver.resolution, resolver.userLoader)
					revision.Save()

					// Save log
					hook := &HookBoltDB{}
					resolver.eventLog.Save(hook)


	/*
			///////////////////// APPLY

			// Apply changes (if emulateDeployment == true --> we set noop to skip deployment part)
			apply := NewEngineApply(diff)
			if !(noop || emulateDeployment) {
				err := apply.Apply()
				apply.SaveLog()
				if err != nil {
					return fmt.Errorf("Cannot apply policy: %v", err)
				}
			}

			// If everything is successful, then increment revision and save run
			// if emulateDeployment == true --> we set noop to false to write state on disk)
			revision := GetLastRevision(GetAptomiBaseDir())
			diff.ProcessSuccessfulExecution(revision, newrevision, noop && !emulateDeployment)
*/
