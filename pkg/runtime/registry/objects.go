package registry

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	// Types represents list of all storable objects
	Types = runtime.AppendAllTypes(engine.Types, lang.PolicyTypes)
)
