package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
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
func NewAttachDependencyAction(componentKey string, dependencyID string) *AttachDependencyAction {
	return &AttachDependencyAction{
		TypeKind:     AttachDependencyActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(AttachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// Apply applies the action
func (a *AttachDependencyAction) Apply(context *action.Context) error {
	context.EventLog.WithFields(event.Fields{
		"componentKey": a.ComponentKey,
		"dependency":   a.DependencyID,
	}).Debug("Attaching dependency '" + a.DependencyID + "' to component instance: " + a.ComponentKey)

	return updateActualStateFromDesired(a.ComponentKey, context, false, false, false)
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *AttachDependencyAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":       a.Kind,
		"key":        a.ComponentKey,
		"dependency": a.DependencyID,
		"pretty":     fmt.Sprintf("[>] %s = %s", a.ComponentKey, a.DependencyID),
	}
}
