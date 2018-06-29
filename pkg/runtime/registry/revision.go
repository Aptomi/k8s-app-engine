package registry

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// GetRevision returns Revision for specified generation
func (reg *defaultRegistry) GetRevision(gen runtime.Generation) (*engine.Revision, error) {
	dataObj, err := reg.store.GetGen(engine.RevisionKey, gen)
	if err != nil {
		return nil, err
	}
	if dataObj == nil {
		return nil, nil
	}

	data, ok := dataObj.(*engine.Revision)
	if !ok {
		return nil, fmt.Errorf("unexpected type while getting Revision from DB")
	}

	return data, nil
}

// NewRevision creates a new revision and saves it to the database
func (reg *defaultRegistry) NewRevision(policyGen runtime.Generation, resolution *resolve.PolicyResolution, recalculateAll bool) (*engine.Revision, error) {
	currRevision, err := reg.GetRevision(runtime.LastGen)
	if err != nil {
		return nil, fmt.Errorf("error while getting last revision: %s", err)
	}

	var gen runtime.Generation
	if currRevision == nil {
		gen = runtime.FirstGen
	} else {
		gen = currRevision.GetGeneration().Next()
	}

	// create revision
	revision := engine.NewRevision(gen, policyGen, recalculateAll)

	// save revision
	_, err = reg.store.Save(revision)
	if err != nil {
		return nil, fmt.Errorf("error while saving new revision: %s", err)
	}

	// save desired state
	desiredState := engine.NewDesiredState(revision, resolution)
	_, err = reg.store.Save(desiredState)
	if err != nil {
		return nil, fmt.Errorf("error while saving desired state for new revision: %s", err)
	}

	return revision, nil
}

// UpdateRevision updates specified Revision in the registry without creating new generation
func (reg *defaultRegistry) UpdateRevision(revision *engine.Revision) error {
	_, err := reg.store.Update(revision)
	if err != nil {
		return fmt.Errorf("error while updating revision: %s", err)
	}

	return nil
}

// GetLastRevisionForPolicy returns last revision for specified policy generation in chronological order
func (reg *defaultRegistry) GetLastRevisionForPolicy(policyGen runtime.Generation) (*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	revisionObjs, err := reg.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	var result *engine.Revision
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision) // nolint: errcheck

		if revision.PolicyGen != policyGen {
			continue
		}

		if result == nil || revision.GetGeneration() > result.GetGeneration() {
			result = revision
		}
	}

	return result, nil
}

// GetAllRevisionsForPolicy returns all revisions for the specified policy generation
func (reg *defaultRegistry) GetAllRevisionsForPolicy(policyGen runtime.Generation) ([]*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	revisionObjs, err := reg.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	result := []*engine.Revision{}
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision) // nolint: errcheck
		if revision.PolicyGen == policyGen {
			result = append(result, revision)
		}
	}

	return result, nil
}

// GetFirstUnprocessedRevision returns the last revision which has not beed processed by the engine yet
func (reg *defaultRegistry) GetFirstUnprocessedRevision() (*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	revisionObjs, err := reg.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	var result *engine.Revision
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision) // nolint: errcheck

		// if this revision has been processed, we don't need to consider it
		if revision.Status == engine.RevisionStatusCompleted || revision.Status == engine.RevisionStatusError {
			continue
		}

		if result == nil || revision.GetGeneration() < result.GetGeneration() {
			result = revision
		}
	}

	return result, nil
}

// GetDesiredState returns desired state associated with the revision
func (reg *defaultRegistry) GetDesiredState(revision *engine.Revision) (*resolve.PolicyResolution, error) {
	obj, err := reg.store.Get(runtime.KeyFromParts(runtime.SystemNS, engine.DesiredStateObject.Kind, engine.GetDesiredStateName(revision.GetGeneration())))
	if err != nil {
		return nil, err
	}
	desiredState, ok := obj.(*engine.DesiredState)
	if !ok {
		return nil, fmt.Errorf("tried to load desired state from the registry, but loaded %v", desiredState)
	}

	return &desiredState.Resolution, nil
}
