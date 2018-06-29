package component

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// AttachClaimActionObject is an informational data structure with Kind and Constructor for the action
var AttachClaimActionObject = &runtime.TypeInfo{
	Kind:        "action-component-claim-attach",
	Constructor: func() runtime.Object { return &AttachClaimAction{} },
}

// AttachClaimAction is a action which gets called when a consumer is added to an existing component
type AttachClaimAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
	ClaimKey     string
	Depth        int
}

// NewAttachClaimAction creates new AttachClaimAction
func NewAttachClaimAction(componentKey string, claimKey string, depth int) *AttachClaimAction {
	return &AttachClaimAction{
		TypeKind:     AttachClaimActionObject.GetTypeKind(),
		Metadata:     action.NewMetadata(AttachClaimActionObject.Kind, componentKey, claimKey),
		ComponentKey: componentKey,
		ClaimKey:     claimKey,
		Depth:        depth,
	}
}

// Apply applies the action
func (a *AttachClaimAction) Apply(context *action.Context) (errResult error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}

		action.CollectMetricsFor(a, start, errResult)
	}()

	context.EventLog.NewEntry().Debugf("Attaching claim '%s' to component instance: '%s'", a.ClaimKey, a.ComponentKey)

	return context.ActualStateUpdater.UpdateComponentInstance(a.ComponentKey, func(obj *resolve.ComponentInstance) {
		obj.ClaimKeys[a.ClaimKey] = a.Depth
	})
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *AttachClaimAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":   a.Kind,
		"key":    a.ComponentKey,
		"claim":  a.ClaimKey,
		"pretty": fmt.Sprintf("[>] %s = %s", a.ComponentKey, a.ClaimKey),
	}
}
