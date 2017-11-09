package store

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	// Objects represents list of all storable objects
	Objects = runtime.AppendAll(engine.Objects, lang.PolicyObjects)
)
