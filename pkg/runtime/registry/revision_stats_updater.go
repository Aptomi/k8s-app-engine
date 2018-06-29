package registry

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/runtime/store"
)

// RevisionResultUpdaterImpl is a default thread-safe implementation of ApplyResultUpdater
type RevisionResultUpdaterImpl struct {
	store    registry.Interface
	revision *engine.Revision
	mutex    sync.Mutex
}

// NewRevisionResultUpdater creates a new default thread-safe implementation of RevisionResultUpdaterImpl, which also
// saves revision on every action
func (ds *defaultStore) NewRevisionResultUpdater(revision *engine.Revision) action.ApplyResultUpdater {
	return &RevisionResultUpdaterImpl{
		store:    ds,
		revision: revision,
	}
}

// SetTotal safely sets the total number of actions
func (updater *RevisionResultUpdaterImpl) SetTotal(total uint32) {
	atomic.StoreUint32(&updater.revision.Result.Total, total)
	updater.revision.Status = engine.RevisionStatusInProgress
	updater.save()
}

// AddSuccess safely increments the number of successfully executed actions
func (updater *RevisionResultUpdaterImpl) AddSuccess() {
	atomic.AddUint32(&updater.revision.Result.Success, 1)
	updater.save()
}

// AddFailed safely increments the number of failed actions
func (updater *RevisionResultUpdaterImpl) AddFailed() {
	atomic.AddUint32(&updater.revision.Result.Failed, 1)
	updater.save()
}

// AddSkipped safely increments the number of skipped actions
func (updater *RevisionResultUpdaterImpl) AddSkipped() {
	atomic.AddUint32(&updater.revision.Result.Skipped, 1)
	updater.save()
}

// Done saves the revision when all actions have been processed
func (updater *RevisionResultUpdaterImpl) Done() *action.ApplyResult {
	if updater.revision.Result.Success+updater.revision.Result.Failed+updater.revision.Result.Skipped != updater.revision.Result.Total {
		panic(fmt.Sprintf("error while applying actions: %d (success) + %d (failed) + %d (skipped) != %d (total)", updater.revision.Result.Success, updater.revision.Result.Failed, updater.revision.Result.Skipped, updater.revision.Result.Total))
	}
	updater.revision.Status = engine.RevisionStatusCompleted
	updater.revision.AppliedAt = time.Now()
	updater.save()
	return updater.revision.Result
}

func (updater *RevisionResultUpdaterImpl) save() {
	updater.mutex.Lock()
	defer updater.mutex.Unlock()
	err := updater.store.UpdateRevision(updater.revision)
	if err != nil {
		panic(fmt.Sprintf("error while saving revision %s: %s", updater.revision.GetGeneration(), err))
	}
}
