package component

import (
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
)

type AttachDependencyAction struct {
	object.Metadata
	*action.Base

	ComponentKey string
	DependencyId string
}

func NewAttachDependencyAction(componentKey string, dependencyId string) *AttachDependencyAction {
	return &AttachDependencyAction{
		Metadata:     object.Metadata{}, // TODO: initialize
		Base:         action.NewBase(),
		ComponentKey: componentKey,
		DependencyId: dependencyId,
	}
}

func (attachDependency *AttachDependencyAction) Apply(context *action.Context) error {
	return nil
}
