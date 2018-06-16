package component

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// DetachDependencyActionObject is an informational data structure with Kind and Constructor for the action
var DetachDependencyActionObject = &runtime.Info{
	Kind:        "action-component-dependency-detach",
	Constructor: func() runtime.Object { return &DetachDependencyAction{} },
}

// DetachDependencyAction is a action which gets called when a consumer is removed from an existing component
type DetachDependencyAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
	DependencyID string
}

// NewDetachDependencyAction creates new DetachDependencyAction
func NewDetachDependencyAction(componentKey string, dependencyID string) *DetachDependencyAction {
	return &DetachDependencyAction{
		Metadata:     action.NewMetadata(DetachDependencyActionObject.Kind, componentKey, dependencyID),
		ComponentKey: componentKey,
		DependencyID: dependencyID,
	}
}

// Apply applies the action
func (a *DetachDependencyAction) Apply(context *action.Context) error {
	context.EventLog.NewEntry().Debugf("Detaching dependency '%s' from component instance: '%s'", a.DependencyID, a.ComponentKey)

	return context.ActualStateUpdater.UpdateComponentInstance(a.ComponentKey, func(obj *resolve.ComponentInstance) {
		delete(obj.DependencyKeys, a.DependencyID)
	})
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *DetachDependencyAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":       a.Kind,
		"key":        a.ComponentKey,
		"dependency": a.DependencyID,
		"pretty":     fmt.Sprintf("[<] %s = %s", a.ComponentKey, a.DependencyID),
	}
}
