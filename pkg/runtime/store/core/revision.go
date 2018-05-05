package core

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
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

// GetFirstRevisionForPolicy returns first revision for specified policy generation in chronological order
func (ds *defaultStore) GetFirstRevisionForPolicy(policyGen runtime.Generation) (*engine.Revision, error) {
	revisionObjs, err := ds.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	var result *engine.Revision
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision)

		if revision.Policy != policyGen {
			continue
		}

		if result == nil || revision.GetGeneration() < result.GetGeneration() {
			result = revision
		}
	}

	return result, nil
}

// GetAllRevisionsForPolicy returns all revisions for the specified policy generation
func (ds *defaultStore) GetAllRevisionsForPolicy(policyGen runtime.Generation) ([]*engine.Revision, error) {
	revisionObjs, err := ds.store.ListGenerations(engine.RevisionKey)
	if err != nil {
		return nil, err
	}

	result := []*engine.Revision{}
	for _, revisionObj := range revisionObjs {
		revision := revisionObj.(*engine.Revision)
		if revision.Policy == policyGen {
			result = append(result, revision)
		}
	}

	return result, nil
}

// NewRevision returns new Revision for specified policy generation
func (ds *defaultStore) NewRevision(policyGen runtime.Generation) (*engine.Revision, error) {
	currRevision, err := ds.GetRevision(runtime.LastGen)
	if err != nil {
		return nil, fmt.Errorf("error while geting current revision: %s", err)
	}

	var gen runtime.Generation
	if currRevision == nil {
		gen = runtime.FirstGen
	} else {
		gen = currRevision.GetGeneration().Next()
	}

	return engine.NewRevision(gen, policyGen), nil
}

// SaveRevision saves specified Revision into the store with possibly new generation creation
func (ds *defaultStore) SaveRevision(revision *engine.Revision) error {
	_, err := ds.store.Save(revision)
	if err != nil {
		return fmt.Errorf("error while saving revision: %s", err)
	}

	return nil
}

// UpdateRevision updates specified Revision in the store without creating new generation
func (ds *defaultStore) UpdateRevision(revision *engine.Revision) error {
	_, err := ds.store.Update(revision)
	if err != nil {
		return fmt.Errorf("error while updating revision: %s", err)
	}

	return nil
}
