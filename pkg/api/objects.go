package api

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

var (
	Objects = runtime.AppendAll([]*runtime.Info{
		EndpointsObject,
		PolicyUpdateResultObject,
		VersionObject,
		ServerErrorObject,
	}, lang.PolicyObjects, engine.Objects)
)
