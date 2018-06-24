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

// DetachClaimActionObject is an informational data structure with Kind and Constructor for the action
var DetachClaimActionObject = &runtime.Info{
	Kind:        "action-component-claim-detach",
	Constructor: func() runtime.Object { return &DetachClaimAction{} },
}

// DetachClaimAction is a action which gets called when a consumer is removed from an existing component
type DetachClaimAction struct {
	runtime.TypeKind `yaml:",inline"`
	*action.Metadata
	ComponentKey string
	ClaimKey     string
}

// NewDetachClaimAction creates new DetachClaimAction
func NewDetachClaimAction(componentKey string, claimKey string) *DetachClaimAction {
	return &DetachClaimAction{
		Metadata:     action.NewMetadata(DetachClaimActionObject.Kind, componentKey, claimKey),
		ComponentKey: componentKey,
		ClaimKey:     claimKey,
	}
}

// Apply applies the action
func (a *DetachClaimAction) Apply(context *action.Context) (errResult error) {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			errResult = fmt.Errorf("panic: %s\n%s", err, string(debug.Stack()))
		}

		action.CollectMetricsFor(a, start, errResult)
	}()

	context.EventLog.NewEntry().Debugf("Detaching claim '%s' from component instance: '%s'", a.ClaimKey, a.ComponentKey)

	return context.ActualStateUpdater.UpdateComponentInstance(a.ComponentKey, func(obj *resolve.ComponentInstance) {
		delete(obj.ClaimKeys, a.ClaimKey)
	})
}

// DescribeChanges returns text-based description of changes that will be applied
func (a *DetachClaimAction) DescribeChanges() util.NestedParameterMap {
	return util.NestedParameterMap{
		"kind":   a.Kind,
		"key":    a.ComponentKey,
		"claim":  a.ClaimKey,
		"pretty": fmt.Sprintf("[<] %s = %s", a.ComponentKey, a.ClaimKey),
	}
}
