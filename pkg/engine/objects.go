package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	// ActionObjects is a list of informational objects for all actions
	ActionObjects = []*runtime.Info{
		component.CreateActionObject,
		component.UpdateActionObject,
		component.DeleteActionObject,
		component.AttachDependencyActionObject,
		component.DetachDependencyActionObject,
		component.EndpointsActionObject,
	}

	// Objects is the list of informational objects for all objects in the engine
	Objects = runtime.AppendAll([]*runtime.Info{
		PolicyDataObject,
		RevisionObject,
		DesiredStateObject,
		resolve.ComponentInstanceObject,
	}, ActionObjects)
)
