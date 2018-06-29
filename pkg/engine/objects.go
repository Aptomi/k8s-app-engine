package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	// Objects is the list of informational objects for all objects in the engine
	Objects = runtime.AppendAllTypes([]*runtime.TypeInfo{
		PolicyDataObject,
		RevisionObject,
		DesiredStateObject,
		resolve.ComponentInstanceObject,
	})
)
