package core

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
	log "github.com/Sirupsen/logrus"
	"math"
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

	return &engine.Revision{
		TypeKind: engine.RevisionObject.GetTypeKind(),
		Metadata: runtime.GenerationMetadata{
			Generation: gen,
		},
		Policy: policyGen,
	}, nil
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

func (ds *defaultStore) GetRevisionProgressUpdater(revision *engine.Revision) progress.Indicator {
	return &revisionProgressUpdater{ds, revision}
}

type revisionProgressUpdater struct {
	store    store.Core
	revision *engine.Revision
}

func (p *revisionProgressUpdater) save() {
	err := p.store.UpdateRevision(p.revision)
	if err != nil {
		log.Warnf("Unable to save revision %s progress with err: %s", p.revision.GetGeneration(), err)
	}
}

func (p *revisionProgressUpdater) SetTotal(total int) {
	p.revision.Progress.Total = total
	p.save()
}

func (p *revisionProgressUpdater) Advance() {
	p.revision.Progress.Current++
	p.save()
}

func (p *revisionProgressUpdater) Done() {
	p.revision.Progress.Current = p.revision.Progress.Total
	p.revision.Progress.Finished = true
	p.save()
}

func (p *revisionProgressUpdater) IsDone() bool {
	return p.revision.Progress.Finished
}

func (p *revisionProgressUpdater) GetCompletionPercent() int {
	if p.revision.Progress.Total <= 0 {
		return 0
	}
	result := int(math.Floor(100.0 * float64(p.revision.Progress.Current) / float64(p.revision.Progress.Total)))
	if result > 100 {
		result = 100
	}
	return result
}
