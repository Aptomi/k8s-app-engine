package api

import (
	"github.com/Aptomi/aptomi/pkg/api/reqresp"
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func handleVersion(r *http.Request, p httprouter.Params) reqresp.Response {
	return version.GetBuildInfo()
}
