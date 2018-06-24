package api

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/version"
)

var (
	// Objects is a list of all objects used in API
	Objects = runtime.AppendAll([]*runtime.Info{
		ClaimsStatusObject,
		PolicyUpdateResultObject,
		AuthSuccessObject,
		AuthRequestObject,
		ServerErrorObject,
		version.BuildInfoObject,
	}, lang.PolicyObjects, engine.Objects)
)
