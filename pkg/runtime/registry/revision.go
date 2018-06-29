package registry

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

// GetRevision returns Revision for specified generation
func (reg *defaultRegistry) GetRevision(gen runtime.Generation) (*engine.Revision, error) {
	// todo thing about replacing hardcoded key with some flag in Info that will show that there is a single object of that kind
	var revision *engine.Revision
	err := reg.store.Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithGen(gen)).One(revision)
	if err != nil {
		return nil, err
	}
	if revision == nil {
		return nil, nil
	}

	return revision, nil
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
	// todo is there a chance that we'll create new revision with all the same data? and Save will not really create new version?
	// todo add WithForceNewVersion?
	err = reg.store.Save(revision)
	if err != nil {
		return nil, fmt.Errorf("error while saving new revision: %s", err)
	}

	// save desired state
	desiredState := engine.NewDesiredState(revision, resolution)
	err = reg.store.Save(desiredState)
	if err != nil {
		return nil, fmt.Errorf("error while saving desired state for new revision: %s", err)
	}

	return revision, nil
}

// UpdateRevision updates specified Revision in the registry without creating new generation
func (reg *defaultRegistry) UpdateRevision(revision *engine.Revision) error {
	// todo add WithInPlace
	err := reg.store.Save(revision)
	if err != nil {
		return fmt.Errorf("error while updating revision: %s", err)
	}

	return nil
}

// GetLastRevisionForPolicy returns last revision for specified policy generation in chronological order
func (reg *defaultRegistry) GetLastRevisionForPolicy(policyGen runtime.Generation) (*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	var revision *engine.Revision
	err := reg.store.Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithWhereEq("PolicyGen", policyGen), store.WithGetLast()).One(revision)
	if err != nil {
		return nil, err
	}

	return revision, nil
}

// GetAllRevisionsForPolicy returns all revisions for the specified policy generation
func (reg *defaultRegistry) GetAllRevisionsForPolicy(policyGen runtime.Generation) ([]*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	var revisions []*engine.Revision
	err := reg.store.Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithWhereEq("PolicyGen", policyGen)).List(&revisions)
	if err != nil {
		return nil, err
	}

	return revisions, nil
}

// GetFirstUnprocessedRevision returns the last revision which has not beed processed by the engine yet
func (reg *defaultRegistry) GetFirstUnprocessedRevision() (*engine.Revision, error) {
	// TODO: this method is slow, needs indexes
	var revision *engine.Revision
	err := reg.store.Find(engine.TypeRevision.Kind, store.WithKey(engine.RevisionKey), store.WithWhereEq("Status", engine.RevisionStatusWaiting, engine.RevisionStatusInProgress), store.WithGetFirst()).One(revision)
	if err != nil {
		return nil, err
	}

	return revision, nil
}

// GetDesiredState returns desired state associated with the revision
func (reg *defaultRegistry) GetDesiredState(revision *engine.Revision) (*resolve.PolicyResolution, error) {
	// todo make desired state versioned same as revision (forceSpecificVersion on save)
	// todo thing about replacing hardcoded key with some flag in Info that will show that there is a single object of that kind
	var desiredState *engine.DesiredState
	err := reg.store.Find(engine.TypeDesiredState.Kind, store.WithKey(runtime.KeyFromParts(runtime.SystemNS, engine.TypeDesiredState.Kind, engine.GetDesiredStateName(revision.GetGeneration())))).One(desiredState)
	if err != nil {
		return nil, err
	}

	return &desiredState.Resolution, nil
}
