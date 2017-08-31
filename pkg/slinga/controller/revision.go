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
	// 1. acquire global lock for revision creation
	// 1. Load last revision
	// 1. Add / match "update" objects => new policy (check in prev revision, map[ns+kind+name]=>object, use deriveEquals?)
	// 1. policyChanged bool
	// 1. run resolver // resolver.ResolveAllDependencies()
	// 1. calculate diff - resolutionChanged bool
	// 1. if policyChanged || resolutionChanged => new Revision, else return no new revision needed
	// 1. RevisionSummary object wraps Revision and stores summary
	// 1. process text policy diff, add to RevisionSummary
	// 1. generate new charts - NewPolicyVisualization, add to RevisionSummary
	// 1. save text diff for component instances into RevisionSummary
	// 1. save new revision
	// 1. always save resolution (event) log even if no new revision created, return event log id to user
	// 1. attach resolution event log id to revision, add to RevisionSummary
	// 1. release global lock for revision creation

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

				// Generate pictures
				visual := NewPolicyVisualization(diff)
				visual.DrawAndStore()

				// Save new resolved state in the last run directory
				resolver.SaveResolutionData()

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
