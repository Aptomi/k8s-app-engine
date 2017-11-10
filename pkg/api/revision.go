package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func (api *coreAPI) handleRevisionGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := params.ByName("gen")

	if len(gen) == 0 {
		gen = strconv.Itoa(int(runtime.LastGen))
	}

	revision, err := api.store.GetRevision(runtime.ParseGeneration(gen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested revision: %s", err))
	}

	api.contentType.Write(writer, request, revision)
}

func (api *coreAPI) handleRevisionGetByPolicy(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	policyGen := params.ByName("policy")

	if len(policyGen) == 0 {
		policyGen = strconv.Itoa(int(runtime.LastGen))
	}

	revision, err := api.store.GetFirstRevisionForPolicy(runtime.ParseGeneration(policyGen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested revision: %s", err))
	}

	api.contentType.Write(writer, request, revision)
}
