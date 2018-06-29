package engine

import (
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	// Types is the list of informational objects for all objects in the engine
	Types = runtime.AppendAllTypes([]*runtime.TypeInfo{
		TypePolicyData,
		TypeRevision,
		TypeDesiredState,
		resolve.TypeComponentInstance,
	})
)
