package actioninfo

import (
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/global"
	"github.com/Aptomi/aptomi/pkg/object"
)

var (
	// Objects is a list of the object.Info for all action types
	Objects = []*object.Info{
		component.CreateActionObject,
		component.UpdateActionObject,
		component.DeleteActionObject,
		component.AttachDependencyActionObject,
		component.DetachDependencyActionObject,
		component.EndpointsActionObject,
		global.PostProcessActionObject,
	}
)
