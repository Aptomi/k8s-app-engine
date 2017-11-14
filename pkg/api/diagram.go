package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/visualization"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type graphWrapper struct {
	Data interface{}
}

func (g *graphWrapper) GetKind() string {
	return "graph"
}

func (api *coreAPI) handlePolicyDiagram(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := params.ByName("gen")

	if len(gen) == 0 {
		gen = strconv.Itoa(int(runtime.LastGen))
	}

	policy, _, err := api.store.GetPolicy(runtime.ParseGeneration(gen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	graph := visualization.NewGraphBuilder(policy, nil, nil).Policy(visualization.PolicyCfgDefault)
	api.contentType.WriteOne(writer, request, &graphWrapper{Data: graph.GetData()})
}

func (api *coreAPI) handlePolicyDiagramCompare(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := params.ByName("gen")
	if len(gen) == 0 {
		gen = strconv.Itoa(int(runtime.LastGen))
	}

	genBase := params.ByName("genBase")
	if len(genBase) == 0 {
		genBase = strconv.Itoa(int(runtime.LastGen))
	}

	policy, _, err := api.store.GetPolicy(runtime.ParseGeneration(gen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}
	policyBase, _, err := api.store.GetPolicy(runtime.ParseGeneration(genBase))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	graph := visualization.NewGraphBuilder(policy, nil, nil).Policy(visualization.PolicyCfgDefault)
	graphBase := visualization.NewGraphBuilder(policyBase, nil, nil).Policy(visualization.PolicyCfgDefault)

	graph.CalcDelta(graphBase)
	api.contentType.WriteOne(writer, request, &graphWrapper{Data: graph.GetData()})
}
