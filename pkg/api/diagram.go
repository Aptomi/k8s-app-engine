package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/visualization"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"strings"
)

type graphWrapper struct {
	Data interface{}
}

func (g *graphWrapper) GetKind() string {
	return "graph"
}

func (api *coreAPI) handlePolicyDiagram(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	mode := params.ByName("mode")
	gen := params.ByName("gen")

	if len(gen) == 0 {
		gen = strconv.Itoa(int(runtime.LastGen))
	}

	policy, _, err := api.store.GetPolicy(runtime.ParseGeneration(gen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	var graph *visualization.Graph
	switch strings.ToLower(mode) {
	case "policy":
		// show just policy
		graphBuilder := visualization.NewGraphBuilder(policy, nil, nil)
		graph = graphBuilder.Policy(visualization.PolicyCfgDefault)
	case "desired":
		// show instances in desired state
		// todo: add request id to the event log scope
		resolver := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog("api-policy-diagram", true))
		state := resolver.ResolveAllDependencies()
		graphBuilder := visualization.NewGraphBuilder(policy, state, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)
	case "actual":
		// show instances in actual state
		state, _ := api.store.GetActualState()
		{
			// since we are not storing dependency keys, calculate them on the fly for actual state
			resolver := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog("api-policy-diagram", true))
			desiredState := resolver.ResolveAllDependencies()
			state.SetDependencyInstanceMap(desiredState.GetDependencyInstanceMap())
		}

		graphBuilder := visualization.NewGraphBuilder(policy, state, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)
	default:
		panic("uknown mode: " + mode)
	}

	api.contentType.WriteOne(writer, request, &graphWrapper{Data: graph.GetData()})
}

func (api *coreAPI) handlePolicyDiagramCompare(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	mode := params.ByName("mode")
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

	var graph *visualization.Graph
	switch strings.ToLower(mode) {
	case "policy":
		// show just policy diff
		graph = visualization.NewGraphBuilder(policy, nil, nil).Policy(visualization.PolicyCfgDefault)
		graphBase := visualization.NewGraphBuilder(policyBase, nil, nil).Policy(visualization.PolicyCfgDefault)
		graph.CalcDelta(graphBase)
	case "desired":
		// show instances in desired state (diff)
		// todo: add request id to the event log scope
		resolver := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog("api-policy-diagram", true))
		state := resolver.ResolveAllDependencies()
		graphBuilder := visualization.NewGraphBuilder(policy, state, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		// todo: add request id to the event log scope
		resolverBase := resolve.NewPolicyResolver(policyBase, api.externalData, event.NewLog("api-policy-diagram", true))
		stateBase := resolverBase.ResolveAllDependencies()
		graphBuilderBase := visualization.NewGraphBuilder(policyBase, stateBase, api.externalData)
		graphBase := graphBuilderBase.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		graph.CalcDelta(graphBase)
	case "actual":
		// show instances in actual state (diff)
		state, _ := api.store.GetActualState()
		{
			// since we are not storing dependency keys, calculate them on the fly for actual state
			resolver := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog("api-policy-diagram", true))
			desiredState := resolver.ResolveAllDependencies()
			state.SetDependencyInstanceMap(desiredState.GetDependencyInstanceMap())
		}

		// show instances in desired state (diff)
		graphBuilder := visualization.NewGraphBuilder(policy, state, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		graphBuilderBase := visualization.NewGraphBuilder(policyBase, state, api.externalData)
		graphBase := graphBuilderBase.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		graph.CalcDelta(graphBase)
	default:
		panic("uknown mode: " + mode)
	}

	api.contentType.WriteOne(writer, request, &graphWrapper{Data: graph.GetData()})
}

func (api *coreAPI) handleObjectDiagram(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	ns := params.ByName("ns")
	kind := params.ByName("kind")
	name := params.ByName("name")

	policy, _, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while getting policy: %s", err))
	}

	obj, err := policy.GetObject(kind, name, ns)
	if err != nil {
		panic(fmt.Sprintf("error while getting object from policy: %s", err))
	}

	var resolution *resolve.PolicyResolution
	if kind == lang.DependencyObject.Kind {
		resolver := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog("api-object-diagram", true))
		resolution = resolver.ResolveAllDependencies()
	}

	var graph *visualization.Graph
	graphBuilder := visualization.NewGraphBuilder(policy, resolution, api.externalData)
	graph = graphBuilder.Object(obj)

	api.contentType.WriteOne(writer, request, &graphWrapper{Data: graph.GetData()})
}
