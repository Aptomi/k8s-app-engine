package api

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/version"
)

var (
	// Types is a list of all objects used in API
	Types = runtime.AppendAllTypes([]*runtime.TypeInfo{
		TypeClaimsStatus,
		TypePolicyUpdateResult,
		TypeAuthSuccess,
		TypeAuthRequest,
		TypeServerError,
		version.TypeBuildInfo,
	}, lang.PolicyTypes, engine.Types)
)
