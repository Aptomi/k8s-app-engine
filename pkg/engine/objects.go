package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/cluster"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/object"
)

var (
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
