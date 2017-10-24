package api

import (
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func handleVersion(r *http.Request, p httprouter.Params) Response {
	return version.GetBuildInfo()
}
