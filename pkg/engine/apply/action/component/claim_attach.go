package component

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/util"
)

// AttachClaimAction is a action which gets called when a consumer is added to an existing component
type AttachClaimAction struct {
	*action.Metadata
	ComponentKey string
	ClaimKey     string
	Depth        int
}

// NewAttachClaimAction creates new AttachClaimAction
func NewAttachClaimAction(componentKey string, claimKey string, depth int) *AttachClaimAction {
	return &AttachClaimAction{
		Metadata:     action.NewMetadata("action-component-claim-attach", componentKey, claimKey),
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
