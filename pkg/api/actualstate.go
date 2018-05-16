package api

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (api *coreAPI) handleActualStateReset(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	err := api.store.ResetActualState()
	if err != nil {
		panic(fmt.Sprintf("error while resetting actual state: %s", err))
	}

	// signal to the channel that policy has changed, that will trigger the enforcement right away
	api.runEnforcement <- true

	api.handleRevisionGet(writer, request, params)
}
