package api

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/version"
)

var (
	// Objects is a list of all objects used in API
	Objects = runtime.AppendAllTypes([]*runtime.TypeInfo{
		TypeClaimsStatus,
		TypePolicyUpdateResult,
		TypeAuthSuccess,
		TypeAuthRequest,
		TypeServerError,
		version.TypeBuildInfo,
	}, lang.PolicyObjects, engine.Objects)
)
