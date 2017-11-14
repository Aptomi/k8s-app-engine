package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/runtime"
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
	actual := context.ActualState.ComponentInstanceMap[a.ComponentKey]
	// in case if create component instance failed or deleted there will be no component instance in actual state
	if actual == nil {
		return nil
	}

	// preserve previous create and update date before overwriting
	desired := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	desired.UpdateTimes(actual.CreatedAt, actual.UpdatedAt)

	context.ActualState.ComponentInstanceMap[a.ComponentKey] = desired

	err := context.ActualStateUpdater.Save(desired)
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}

	return nil
}
