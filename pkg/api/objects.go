package api

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	// Objects is a list of all objects used in API
	Objects = runtime.AppendAll([]*runtime.Info{
		EndpointsObject,
		PolicyUpdateResultObject,
		VersionObject,
		ServerErrorObject,
	}, lang.PolicyObjects, engine.Objects)
)
