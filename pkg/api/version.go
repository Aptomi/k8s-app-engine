package api

import (
	"net/http"

	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/julienschmidt/httprouter"
)

func (api *coreAPI) handleVersion(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	api.contentType.WriteOne(writer, request, version.GetBuildInfo())
}
