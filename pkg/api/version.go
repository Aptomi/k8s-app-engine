package api

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// VersionObject is an informational data structure with Kind and Constructor for Version
var VersionObject = &runtime.Info{
	Kind:        "version",
	Constructor: func() runtime.Object { return &Version{} },
}

// Version represents build info in the API
type Version struct {
	runtime.TypeKind  `yaml:",inline"`
	version.BuildInfo `yaml:",inline"`
}

var currentVersion = &Version{
	TypeKind:  VersionObject.GetTypeKind(),
	BuildInfo: version.GetBuildInfo(),
}

func (api *coreAPI) handleVersion(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	api.contentType.WriteOne(writer, request, currentVersion)
}
