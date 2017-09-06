package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type DetachDependencyComponent struct {
	object.Metadata
	*action.Base

	ComponentKey string
	DependencyId string
}

func NewDetachDependencyAction(componentKey string, dependencyId string) *DetachDependencyComponent {
	return &DetachDependencyComponent{
		Metadata:     object.Metadata{}, // TODO: initialize
		Base:         action.NewBase(),
		ComponentKey: componentKey,
		DependencyId: dependencyId,
	}
}

func (detachDependency *DetachDependencyComponent) Apply(context *action.Context) error {
	return nil
}
