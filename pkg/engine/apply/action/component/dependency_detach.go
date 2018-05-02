package component

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
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
func NewDetachDependencyAction(componentKey string, dependencyID string) *DetachDependencyAction {
	return &DetachDependencyAction{
		Metadata:     action.NewMetadata(DetachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// Apply applies the action
func (a *DetachDependencyAction) Apply(context *action.Context) error {
	context.EventLog.WithFields(event.Fields{
		"componentKey": a.ComponentKey,
		"dependency":   a.DependencyID,
	}).Debug("Detaching dependency '" + a.DependencyID + "' from component instance: " + a.ComponentKey)

	// if an instance is still present in desired state (haven't been fully deleted due to other dependencies), then update actual state
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	if instance != nil {
		return updateActualStateFromDesired(a.ComponentKey, context, false, false, false)
	}
	return nil
}
