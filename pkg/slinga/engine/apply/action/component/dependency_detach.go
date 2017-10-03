package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

// DetachDependencyActionObject is an informational data structure with Kind and Constructor for the action
var DetachDependencyActionObject = &object.Info{
	Kind:        "action-component-dependency-detach",
	Constructor: func() object.Base { return &DetachDependencyAction{} },
}

// DetachDependencyAction is a action which gets called when a consumer is removed from an existing component
type DetachDependencyAction struct {
	*action.Metadata
	ComponentKey string
	DependencyID string
}

// NewDetachDependencyAction creates new DetachDependencyAction
func NewDetachDependencyAction(revision object.Generation, componentKey string, dependencyID string) *DetachDependencyAction {
	return &DetachDependencyAction{
		Metadata:     action.NewMetadata(revision, DetachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// GetName returns action name
func (a *DetachDependencyAction) GetName() string {
	return "Component " + a.ComponentKey + " detach dependency " + a.DependencyID
}

// Apply applies the action
func (a *DetachDependencyAction) Apply(context *action.Context) error {
	return nil
}
