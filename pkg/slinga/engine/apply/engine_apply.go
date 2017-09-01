package apply

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/plugin"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/progress"
	. "github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

type EngineApply struct {
	// Diff to be applied (as well as next/prev policy & user loader)
	diff       *diff.PolicyResolutionDiff
	nextPolicy *language.PolicyNamespace
	prevPolicy *language.PolicyNamespace
	userLoader language.UserLoader

	// Buffered event log - gets populated while applying changes
	eventLog *EventLog

	// Plugins to execute
	plugins []plugin.EnginePlugin

	// Progress indicator
	progress progress.ProgressIndicator
}

func NewEngineApply(diff *diff.PolicyResolutionDiff, nextPolicy *language.PolicyNamespace, prevPolicy *language.PolicyNamespace, userLoader language.UserLoader) *EngineApply {
	return &EngineApply{
		diff:       diff,
		nextPolicy: nextPolicy,
		prevPolicy: prevPolicy,
		userLoader: userLoader,
		eventLog:   NewEventLog(),
		plugins:    plugin.AllPlugins(),
		progress:   progress.NewProgressConsole(),
	}
}

// Returns difference length (used for progress indicator)
func (apply *EngineApply) GetApplyProgressLength() int {
	result := len(apply.diff.Actions)
	for _, pluginInstance := range apply.plugins {
		result += pluginInstance.GetCustomApplyProgressLength()
	}
	return result
}

// Apply method applies all changes via plugins
func (apply *EngineApply) Apply() error {
	// initialize all plugins
	for _, pluginInstance := range apply.plugins {
		pluginInstance.Init(
			apply.nextPolicy,
			apply.diff.Next,
			apply.prevPolicy,
			apply.diff.Prev,
			apply.userLoader,
			apply.eventLog,
		)
	}

	// initialize progress indicator
	apply.progress.SetTotal(apply.GetApplyProgressLength())

	// error count while applying changes
	foundErrors := false

	// process all actions
	for _, action := range apply.diff.Actions {
		apply.progress.Advance("Action")
		err := action.Apply(apply.plugins, apply.eventLog)
		if err != nil {
			foundErrors = true
		}
	}

	// call plugins to perform their custom apply actions
	for _, pluginInstance := range apply.plugins {
		err := pluginInstance.OnApplyCustom(apply.progress)
		if err != nil {
			apply.eventLog.LogError(fmt.Errorf("Error while calling OnApplyCustom() on a plugin: " + err.Error()))
			foundErrors = true
		}
	}

	// Finalize progress indicator
	apply.progress.Done()

	if foundErrors {
		err := fmt.Errorf("One or more errors occured while applying policy")
		apply.eventLog.LogError(err)
		return err
	}
	return nil
}

func (apply *EngineApply) SaveLog() {
	// Save log
	hook := &HookBoltDB{}
	apply.eventLog.Save(hook)
}
