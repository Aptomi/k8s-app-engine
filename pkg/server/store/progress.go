package store

import (
	"github.com/Aptomi/aptomi/pkg/engine/progress"
	"github.com/Aptomi/aptomi/pkg/object/store"
	log "github.com/Sirupsen/logrus"
	"math"
)

func (s *defaultStore) Progress(store store.ObjectStore, revision *RevisionData) progress.Indicator {
	return &revisionProgressStore{store, revision}
}

type revisionProgress struct {
	Stage    string
	Current  int
	Total    int
	Finished bool
}

type revisionProgressStore struct {
	store    store.ObjectStore
	revision *RevisionData
}

func (p *revisionProgressStore) save() {
	_, err := p.store.Save(p.revision)
	if err != nil {
		log.Panicf("Unable to save revision %s progress with err: %s", p.revision.Generation, err)
	}
}

func (p *revisionProgressStore) SetTotal(total int) {
	p.revision.Progress.Total = total
	p.Advance("Init")
	p.save()
}

func (p *revisionProgressStore) Advance(stage string) {
	p.revision.Progress.Stage = stage
	p.revision.Progress.Current++
	p.save()
}

func (p *revisionProgressStore) Done() {
	p.revision.Progress.Current = p.revision.Progress.Total
	p.revision.Progress.Finished = true
	p.save()
}

func (p *revisionProgressStore) IsDone() bool {
	return p.revision.Progress.Finished
}

func (p *revisionProgressStore) GetCompletionPercent() int {
	if p.revision.Progress.Total <= 0 {
		return 0
	}
	result := int(math.Floor(100.0 * float64(p.revision.Progress.Current) / float64(p.revision.Progress.Total)))
	if result > 100 {
		result = 100
	}
	return result
}
