package api

import (
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (api *coreAPI) handleVersion(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	api.contentType.WriteOne(writer, request, version.GetBuildInfo())
}
