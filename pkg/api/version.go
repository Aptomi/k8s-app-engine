package api

import (
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var VersionObject = &runtime.Info{
	Kind:        "version",
	Constructor: func() runtime.Object { return &Version{} },
}

type Version struct {
	runtime.TypeKind  `yaml:",inline"`
	version.BuildInfo `yaml:",inline"`
}

var currentVersion = &Version{
	TypeKind:  VersionObject.GetTypeKind(),
	BuildInfo: version.GetBuildInfo(),
}

func (api *coreApi) handleVersion(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	api.contentType.Write(writer, request, currentVersion)
}
