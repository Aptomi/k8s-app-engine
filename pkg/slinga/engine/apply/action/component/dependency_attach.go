package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var AttachDependencyActionObject = &object.Info{
	Kind:        "action-component-dependency-attach",
	Constructor: func() object.Base { return &AttachDependencyAction{} },
}

type AttachDependencyAction struct {
	*action.Metadata
	ComponentKey string
	DependencyID string
}

func NewAttachDependencyAction(revision object.Generation, componentKey string, dependencyID string) *AttachDependencyAction {
	return &AttachDependencyAction{
		Metadata:     action.NewMetadata(revision, AttachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

func (a *AttachDependencyAction) GetName() string {
	return "Component " + a.ComponentKey + " attach dependency " + a.DependencyID
}

func (a *AttachDependencyAction) Apply(context *action.Context) error {
	return nil
}
