package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type DetachDependencyAction struct {
	object.Metadata
	*action.Base

	ComponentKey string
	DependencyId string
}

func NewDetachDependencyAction(componentKey string, dependencyId string) *DetachDependencyAction {
	return &DetachDependencyAction{
		Metadata:     object.Metadata{}, // TODO: initialize
		Base:         action.NewBase(),
		ComponentKey: componentKey,
		DependencyId: dependencyId,
	}
}

func (a *DetachDependencyAction) Apply(context *action.Context) error {
	return nil
}
