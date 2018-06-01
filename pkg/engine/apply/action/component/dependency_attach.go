package component

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
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

// AfterCreated allows to modify actual state after an action has been created and added to the tree of actions, but before it got executed
func (a *AttachDependencyAction) AfterCreated(actualState *resolve.PolicyResolution) {

}

// Apply applies the action
func (a *AttachDependencyAction) Apply(context *action.Context) error {
	context.EventLog.NewEntry().Debugf("Attaching dependency '%s' to component instance: '%s'", a.DependencyID, a.ComponentKey)

	// add reference to dependency into the actual state
	instance := context.ActualState.ComponentInstanceMap[a.ComponentKey]
	instance.DependencyKeys[a.DependencyID] = true

	return updateComponentInActualState(a.ComponentKey, context)
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
