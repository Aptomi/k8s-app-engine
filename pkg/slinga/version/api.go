package version

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/slinga/util"
	"net/http"
)

func handleVersion(w http.ResponseWriter, _ *http.Request) {
	err := util.WriteJSON(w, GetBuildInfo())
	if err != nil {
		panic(fmt.Sprintf("Error while serializing BuildInfo: %s", err))
	}
}

func Serve(r *http.ServeMux) {
	r.HandleFunc("/version", handleVersion)
	r.HandleFunc("/version/", handleVersion)
}
