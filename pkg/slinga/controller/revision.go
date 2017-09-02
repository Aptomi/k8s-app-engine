package controller

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/object/store"
)

type RevisionController interface {
	GetRevision(object.Generation) (*language.PolicyNamespace, error)
	NewRevision([]object.Base) error
}

func NewRevisionController(store store.ObjectStore) RevisionController {
	return &RevisionControllerImpl{store}
}

type RevisionControllerImpl struct {
	store store.ObjectStore
}

func (c *RevisionControllerImpl) GetRevision(gen object.Generation) (*language.PolicyNamespace, error) {
	return nil, nil
}

func (c *RevisionControllerImpl) NewRevision(update []object.Base) error {
	// When user submits a change
	// [1] Acquire global lock for policy version creation
	//   1. defer release lock
	//
	// [2] Create new version of policy, if needed
	//   1. add / match "update" objects => new policy (check in prev revision, map[ns+kind+name]=>object, use deriveEquals?)
	//   1. policyChanged bool = calculate if policy changed
	//   1. if policyChanged =>
	// 				save new version of policy
	//				attach resolution event log to the new version of policy
	//      else return "no changes in policy"
	//
	// [3] Show user what changes will be triggered by his changes to the policy
	//   1. load previous desired state (from last revision)
	//   1. calculate new desired state (run resolver // resolver.ResolveAllDependencies())
	//   1. compare and return changes to the user [without saving to db]

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

				policy := ... NewPolicyNamespace()

				resolver := resolve.NewPolicyResolver(policy, userLoader)
				nextState, err := resolver.ResolveAllDependencies()
				if err != nil {
					log.Panicf("Cannot resolve policy: %v", err)
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
						log.Panicf("Cannot apply policy: %v", err)
					}
				}

				// If everything is successful, then increment revision and save run
				// if emulateDeployment == true --> we set noop to false to write state on disk)
				revision := GetLastRevision(GetAptomiBaseDir())
				diff.ProcessSuccessfulExecution(revision, newrevision, noop && !emulateDeployment)
	*/

	return nil
}

/*

func (reg *Registry) LoadPolicy(gen Generation) (*PolicyNamespace, error) {
	policyObj, err := reg.store.GetNewestOne("system", PolicyNamespaceDataObject.Kind, "main")
	if err != nil {
		return nil, err
	}
	policyData, ok := policyObj.(*PolicyNamespaceData)
	if !ok {
		return nil, fmt.Errorf("Can't cast object from store to PolicyData: %v", policyObj)
	}

	policy := NewPolicyNamespace()

	keys := make([]Key, 0, len(policyData.Objects))
	for _, key := range policyData.Objects {
		keys = append(keys, key)
	}

	objects, err := reg.store.GetManyByKeys(keys)
	if err != nil {
		return nil, fmt.Errorf("Can't load objects for policy data %s: %s", policyData.GetKey(), err)
	}

	for _, obj := range objects {
		fmt.Println("Loaded object")
		fmt.Println(obj)
		policy.AddObject(obj)
	}

	return policy, nil
}

*/
