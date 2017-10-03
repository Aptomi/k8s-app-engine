package version

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func handleVersion(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	err := util.WriteJSON(w, GetBuildInfo())
	if err != nil {
		panic(fmt.Sprintf("Error while serializing BuildInfo: %s", err))
	}
}

// Serve registers version handler in the API
func Serve(r *httprouter.Router) {
	r.GET("/version", handleVersion)
}
