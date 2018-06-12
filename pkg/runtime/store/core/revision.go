package core

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/external"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/sirupsen/logrus"
)

// GetRevision returns Revision for specified generation
func (ds *defaultStore) GetRevision(gen runtime.Generation) (*engine.Revision, error) {
	dataObj, err := ds.store.GetGen(engine.RevisionKey, gen)
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
// TODO: this method should save desiredState for this revision in the store (https://github.com/Aptomi/aptomi/issues/318)
func (ds *defaultStore) NewRevision(policyGen runtime.Generation, desiredState *resolve.PolicyResolution, recalculateAll bool) (*engine.Revision, error) {
	currRevision, err := ds.GetRevision(runtime.LastGen)
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
	_, err = ds.store.Save(revision)
	if err != nil {
		return nil, fmt.Errorf("error while saving new revision: %s", err)
	}

	return revision, nil
}

// UpdateRevision updates specified Revision in the store without creating new generation
func (ds *defaultStore) UpdateRevision(revision *engine.Revision) error {
	_, err := ds.store.Update(revision)
	if err != nil {
		return fmt.Errorf("error while updating revision: %s", err)
	}

	return nil
}

// GetLastRevisionForPolicy returns last revision for specified policy generation in chronological order
func (ds *defaultStore) GetLastRevisionForPolicy(policyGen runtime.Generation) (*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	revisionObjs, err := ds.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	var result *engine.Revision
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision)

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
func (ds *defaultStore) GetAllRevisionsForPolicy(policyGen runtime.Generation) ([]*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	revisionObjs, err := ds.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	result := []*engine.Revision{}
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision)
		if revision.PolicyGen == policyGen {
			result = append(result, revision)
		}
	}

	return result, nil
}

// GetFirstUnprocessedRevision returns the last revision which has not beed processed by the engine yet
func (ds *defaultStore) GetFirstUnprocessedRevision() (*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	revisionObjs, err := ds.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	var result *engine.Revision
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision)

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
// TODO: policy and external data need to be removed from the signature of this method, once it starts loading desired state from revision instead of calculating it on the fly https://github.com/Aptomi/aptomi/issues/318
func (ds *defaultStore) GetDesiredState(revision *engine.Revision, policy *lang.Policy, externalData *external.Data) (*resolve.PolicyResolution, error) {
	return resolve.NewPolicyResolver(policy, externalData, event.NewLog(logrus.DebugLevel, fmt.Sprintf("revision-%d-desired-state", revision.GetGeneration()))).ResolveAllDependencies(), nil
}
