package actions

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type ComponentAttachDependency struct {
	object.Metadata
	*BaseAction

	ComponentKey string
	DependencyId string
}

func NewComponentAttachDependencyAction(componentKey string, dependencyId string) *ComponentAttachDependency {
	return &ComponentAttachDependency{
		Metadata:     object.Metadata{}, // TODO: initialize
		BaseAction:   NewComponentBaseAction(),
		ComponentKey: componentKey,
		DependencyId: dependencyId,
	}
}

func (attachDependency *ComponentAttachDependency) Apply(context *ActionContext) error {
	return nil
}
