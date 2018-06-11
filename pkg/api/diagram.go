package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/visualization"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
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
		desiredState := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-policy-diagram")).ResolveAllDependencies()
		graphBuilder := visualization.NewGraphBuilder(policy, desiredState, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)
	case "actual":
		// show instances in actual state
		state, _ := api.store.GetActualState()
		{
			// since we are not storing dependency keys, calculate them on the fly for actual state
			desiredState := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-policy-diagram")).ResolveAllDependencies()
			state.SetDependencyInstanceMap(desiredState.GetDependencyInstanceMap())
		}

		graphBuilder := visualization.NewGraphBuilder(policy, state, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)
	default:
		panic("unknown mode: " + mode)
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
		// policy & policy base
		graph = visualization.NewGraphBuilder(policy, nil, nil).Policy(visualization.PolicyCfgDefault)
		graphBase := visualization.NewGraphBuilder(policyBase, nil, nil).Policy(visualization.PolicyCfgDefault)

		// diff
		graph.CalcDelta(graphBase)
	case "desired":
		// desired state (next)
		desiredState := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-policy-diagram")).ResolveAllDependencies()
		graphBuilder := visualization.NewGraphBuilder(policy, desiredState, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		// desired state (prev)
		desiredStateBase := resolve.NewPolicyResolver(policyBase, api.externalData, event.NewLog(logrus.WarnLevel, "api-policy-diagram")).ResolveAllDependencies()
		graphBuilderBase := visualization.NewGraphBuilder(policyBase, desiredStateBase, api.externalData)
		graphBase := graphBuilderBase.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		// diff
		graph.CalcDelta(graphBase)
	case "actual":
		// actual state
		actualState, _ := api.store.GetActualState()
		{
			// since we are not storing dependency keys, calculate them on the fly for actual state
			desiredState := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-policy-diagram")).ResolveAllDependencies()
			actualState.SetDependencyInstanceMap(desiredState.GetDependencyInstanceMap())
		}

		// desired state
		graphBuilder := visualization.NewGraphBuilder(policy, actualState, api.externalData)
		graph = graphBuilder.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		graphBuilderBase := visualization.NewGraphBuilder(policyBase, actualState, api.externalData)
		graphBase := graphBuilderBase.DependencyResolution(visualization.DependencyResolutionCfgDefault)

		// diff
		graph.CalcDelta(graphBase)
	default:
		panic("unknown mode: " + mode)
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

	var desiredState *resolve.PolicyResolution
	if kind == lang.DependencyObject.Kind {
		desiredState = resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-object-diagram")).ResolveAllDependencies()
	}

	var graph *visualization.Graph
	graphBuilder := visualization.NewGraphBuilder(policy, desiredState, api.externalData)
	graph = graphBuilder.Object(obj)

	api.contentType.WriteOne(writer, request, &graphWrapper{Data: graph.GetData()})
}
