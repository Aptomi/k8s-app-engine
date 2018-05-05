package action

import (
	"fmt"
	"sync/atomic"
)

// ApplyResult is a result of applying actions
type ApplyResult struct {
	Success uint32
	Failed  uint32
	Skipped uint32
	Total   uint32
}

// ApplyResultUpdater is an interface for handling revision progress stats (# of processed actions) when applying action plan
type ApplyResultUpdater interface {
	SetTotal(actions uint32)
	AddSuccess()
	AddFailed()
	AddSkipped()
	Done() *ApplyResult
}

// ApplyResultUpdaterImpl is a default thread-safe implementation of ApplyResultUpdater
type ApplyResultUpdaterImpl struct {
	Result *ApplyResult
}

// NewApplyResultUpdaterImpl creates a new default thread-safe implementation ApplyResultUpdaterImpl of ApplyResultUpdater
func NewApplyResultUpdaterImpl() *ApplyResultUpdaterImpl {
	return &ApplyResultUpdaterImpl{
		Result: &ApplyResult{},
	}
}

// SetTotal safely sets the total number of actions
func (updater *ApplyResultUpdaterImpl) SetTotal(total uint32) {
	atomic.StoreUint32(&updater.Result.Total, total)
}

// AddSuccess safely increments the number of successfully executed actions
func (updater *ApplyResultUpdaterImpl) AddSuccess() {
	atomic.AddUint32(&updater.Result.Success, 1)
}

// AddFailed safely increments the number of failed actions
func (updater *ApplyResultUpdaterImpl) AddFailed() {
	atomic.AddUint32(&updater.Result.Failed, 1)
}

// AddSkipped safely increments the number of skipped actions
func (updater *ApplyResultUpdaterImpl) AddSkipped() {
	atomic.AddUint32(&updater.Result.Skipped, 1)
}

// Done does nothing except doing an integrity check for default implementation
func (updater *ApplyResultUpdaterImpl) Done() *ApplyResult {
	if updater.Result.Success+updater.Result.Failed+updater.Result.Skipped != updater.Result.Total {
		panic(fmt.Sprintf("error while applying actions: %d (success) + %d (failed) + %d (skipped) != %d (total)", updater.Result.Success, updater.Result.Failed, updater.Result.Skipped, updater.Result.Total))
	}
	return updater.Result
}
