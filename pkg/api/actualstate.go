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
}
