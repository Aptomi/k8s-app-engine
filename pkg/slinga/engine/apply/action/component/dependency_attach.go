package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

// AttachDependencyActionObject is an informational data structure with Kind and Constructor for the action
var AttachDependencyActionObject = &object.Info{
	Kind:        "action-component-dependency-attach",
	Constructor: func() object.Base { return &AttachDependencyAction{} },
}

// AttachDependencyAction is a action which gets called when a consumer is added to an existing component
type AttachDependencyAction struct {
	*action.Metadata
	ComponentKey string
	DependencyID string
}

// NewAttachDependencyAction creates new AttachDependencyAction
func NewAttachDependencyAction(revision object.Generation, componentKey string, dependencyID string) *AttachDependencyAction {
	return &AttachDependencyAction{
		Metadata:     action.NewMetadata(revision, AttachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// GetName returns action name
func (a *AttachDependencyAction) GetName() string {
	return "Component " + a.ComponentKey + " attach dependency " + a.DependencyID
}

// Apply applies the action
func (a *AttachDependencyAction) Apply(context *action.Context) error {
	return nil
}
