package component

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/object"
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

// Apply applies the action
func (a *AttachDependencyAction) Apply(context *action.Context) error {
	return nil
}
