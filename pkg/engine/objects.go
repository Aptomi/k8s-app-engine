package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/global"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	ActionObjects = []*runtime.Info{
		component.CreateActionObject,
		component.UpdateActionObject,
		component.DeleteActionObject,
		component.AttachDependencyActionObject,
		component.DetachDependencyActionObject,
		component.EndpointsActionObject,
		global.PostProcessActionObject,
	}
	Objects = runtime.AppendAll([]*runtime.Info{
		PolicyDataObject,
		RevisionObject,
		resolve.ComponentInstanceObject,
	}, ActionObjects)
)
