package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type ComponentDetachDependency struct {
	object.Metadata
	*BaseAction

	ComponentKey string
	DependencyId string
}

func NewComponentDetachDependencyAction(componentKey string, dependencyId string) *ComponentDetachDependency {
	return &ComponentDetachDependency{
		Metadata:     object.Metadata{}, // TODO: initialize
		BaseAction:   NewComponentBaseAction(),
		ComponentKey: componentKey,
		DependencyId: dependencyId,
	}
}

func (detachDependency *ComponentDetachDependency) Apply(context *ActionContext) error {
	return nil
}
