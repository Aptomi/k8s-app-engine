package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/cluster"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/object"
)

var (
	// ActionObjects is a list of the object.Info for all action types
	ActionObjects = []*object.Info{
		component.CreateActionObject,
		component.UpdateActionObject,
		component.DeleteActionObject,
		component.AttachDependencyActionObject,
		component.DetachDependencyActionObject,
		component.EndpointsActionObject,
		cluster.PostProcessActionObject,
	}
)
