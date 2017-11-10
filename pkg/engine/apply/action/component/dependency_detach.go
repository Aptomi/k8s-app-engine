package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"time"
)

// DetachDependencyActionObject is an informational data structure with Kind and Constructor for the action
var DetachDependencyActionObject = &runtime.Info{
	Kind:        "action-component-dependency-detach",
	Constructor: func() runtime.Object { return &DetachDependencyAction{} },
}

// DetachDependencyAction is a action which gets called when a consumer is removed from an existing component
type DetachDependencyAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
	DependencyID string
}

// NewDetachDependencyAction creates new DetachDependencyAction
func NewDetachDependencyAction(revision runtime.Generation, componentKey string, dependencyID string) *DetachDependencyAction {
	return &DetachDependencyAction{
		Metadata:     action.NewMetadata(revision, DetachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// Apply applies the action
func (a *DetachDependencyAction) Apply(context *action.Context) error {
	return a.updateActualState(context)
}

func (a *DetachDependencyAction) updateActualState(context *action.Context) error {
	// preserve previous creation date before overwriting
	prevCreatedOn := context.ActualState.ComponentInstanceMap[a.ComponentKey].CreatedOn
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	instance.UpdateTimes(prevCreatedOn, time.Now())

	context.ActualState.ComponentInstanceMap[a.ComponentKey] = instance
	err := context.ActualStateUpdater.Save(instance)
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}
