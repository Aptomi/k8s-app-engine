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

	if revision == nil {
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusNotFound)
	} else {
		api.contentType.WriteOne(writer, request, revision)
	}
}

type revisionsWrapper struct {
	Data interface{}
}

func (g *revisionsWrapper) GetKind() string {
	return "revisions"
}

func (api *coreAPI) handleRevisionsGetByPolicy(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	policyGen := params.ByName("policy")

	if len(policyGen) == 0 {
		policyGen = strconv.Itoa(int(runtime.LastGen))
	}

	revisions, err := api.store.GetAllRevisionsForPolicy(runtime.ParseGeneration(policyGen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested revisions: %s", err))
	}

	if revisions == nil {
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusNotFound)
	} else {
		api.contentType.WriteOne(writer, request, &revisionsWrapper{Data: revisions})
	}
}
