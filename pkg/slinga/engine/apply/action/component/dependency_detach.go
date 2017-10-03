package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

var DetachDependencyActionObject = &object.Info{
	Kind:        "action-component-dependency-detach",
	Constructor: func() object.Base { return &DetachDependencyAction{} },
}

type DetachDependencyAction struct {
	*action.Metadata
	ComponentKey string
	DependencyID string
}

func NewDetachDependencyAction(revision object.Generation, componentKey string, dependencyID string) *DetachDependencyAction {
	return &DetachDependencyAction{
		Metadata:     action.NewMetadata(revision, DetachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

func (a *DetachDependencyAction) GetName() string {
	return "Component " + a.ComponentKey + " detach dependency " + a.DependencyID
}

func (a *DetachDependencyAction) Apply(context *action.Context) error {
	return nil
}
