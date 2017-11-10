package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"time"
)

// AttachDependencyActionObject is an informational data structure with Kind and Constructor for the action
var AttachDependencyActionObject = &runtime.Info{
	Kind:        "action-component-dependency-attach",
	Constructor: func() runtime.Object { return &AttachDependencyAction{} },
}

// AttachDependencyAction is a action which gets called when a consumer is added to an existing component
type AttachDependencyAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
	DependencyID string
}

// NewAttachDependencyAction creates new AttachDependencyAction
func NewAttachDependencyAction(revision runtime.Generation, componentKey string, dependencyID string) *AttachDependencyAction {
	return &AttachDependencyAction{
		TypeKind:     AttachDependencyActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(revision, AttachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// Apply applies the action
func (a *AttachDependencyAction) Apply(context *action.Context) error {
	return a.updateActualState(context)
}

func (a *AttachDependencyAction) updateActualState(context *action.Context) error {
	// preserve previous creation date before overwriting
	prevCreatedOn := context.ActualState.ComponentInstanceMap[a.ComponentKey].CreatedOn
	instance := context.DesiredState.ComponentInstanceMap[a.ComponentKey]
	instance.UpdateTimes(prevCreatedOn, time.Now())

	context.ActualState.ComponentInstanceMap[a.ComponentKey] = instance
	err := context.ActualStateUpdater.Save(instance)
	if err != nil {
		return fmt.Errorf("error while update actual state: %s", err)
	}
	return nil
}
