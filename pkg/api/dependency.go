package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

// DependencyQueryFlag determines whether to query just dependency deployment status, or both deployment + readiness/health checks status
type DependencyQueryFlag string

const (
	// DependencyQueryDeploymentStatusOnly prescribes only to query dependency deployment status (i.e. actual state = desired state)
	DependencyQueryDeploymentStatusOnly DependencyQueryFlag = "deployed"

	// DependencyQueryDeploymentStatusAndReadiness prescribes to query both dependency deployment status (i.e. actual state = desired state), as well as readiness status (i.e. health checks = passing)
	DependencyQueryDeploymentStatusAndReadiness DependencyQueryFlag = "ready"
)

// DependencyStatusObject is an informational data structure with Kind and Constructor for Dependency Status
var DependencyStatusObject = &runtime.Info{
	Kind:        "dependency-status",
	Constructor: func() runtime.Object { return &DependencyStatus{} },
}

// DependencyStatus is a struct which holds status information for a set of given dependencies
type DependencyStatus struct {
	runtime.TypeKind `yaml:",inline"`

	// map containing status by dependency
	Status map[string]*DependencyStatusIndividual
}

// DependencyStatusIndividual is a struct which holds status information for an individual dependency
type DependencyStatusIndividual struct {
	Found     bool
	Deployed  bool
	Ready     bool
	Endpoints map[string]map[string]string
}

func (api *coreAPI) handleDependencyStatusGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// parse query mode flag (deployment status vs. readiness status) as well as the list of dependency IDs
	flag := DependencyQueryFlag(params.ByName("queryFlag"))
	dependencyIds := strings.Split(params.ByName("idList"), ",")

	// load the latest policy
	policy, _, errPolicy := api.store.GetPolicy(runtime.LastGen)
	if errPolicy != nil {
		panic(fmt.Sprintf("error while loading latest policy from the store: %s", errPolicy))
	}

	// initialize result
	result := &DependencyStatus{
		TypeKind: DependencyStatusObject.GetTypeKind(),
		Status:   make(map[string]*DependencyStatusIndividual),
	}
	for _, depID := range dependencyIds {
		parts := strings.Split(depID, "^")
		dObj, err := policy.GetObject(lang.DependencyObject.Kind, parts[1], parts[0])
		if dObj == nil || err != nil {
			dKey := runtime.KeyFromParts(parts[0], lang.DependencyObject.Kind, parts[1])
			result.Status[dKey] = &DependencyStatusIndividual{
				Found:     false,
				Deployed:  false,
				Ready:     false,
				Endpoints: make(map[string]map[string]string),
			}
			continue
		}

		d := dObj.(*lang.Dependency)
		result.Status[runtime.KeyForStorable(d)] = &DependencyStatusIndividual{
			Found:     true,
			Deployed:  true,
			Ready:     false, // TODO: this needs to be implemented. see https://github.com/Aptomi/aptomi/issues/315
			Endpoints: make(map[string]map[string]string),
		}
	}

	// load actual and desired states
	desiredState := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-dependency-status", false)).ResolveAllDependencies()
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("can't load actual state from the store: %s", err))
	}

	// fetch deployment status for dependencies
	fetchDeploymentStatusForDependencies(result, actualState, desiredState)

	// fetch readiness status for dependencies, if we were asked to do so
	if flag == DependencyQueryDeploymentStatusAndReadiness {
		fetchReadinessStatusForDependencies(result, actualState, desiredState)
	}

	// fetch endpoints for dependencies
	fetchEndpointsForDependencies(result, actualState)

	// return the result back
	api.contentType.WriteOne(writer, request, result)
}

func fetchDeploymentStatusForDependencies(result *DependencyStatus, actualState *resolve.PolicyResolution, desiredState *resolve.PolicyResolution) {
	// compare desired vs. actual state and see what's the dependency status for every provided dependency ID
	diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan.Apply(
		action.WrapSequential(func(act action.Base) error {
			key, ok := act.DescribeChanges()["key"].(string)
			if ok && len(key) > 0 {
				// we found a component in diff, which is affected by the action. let's see if any of the dependencies are affected
				affectedDepKeys := make(map[string]bool)
				{
					prevInstance := actualState.ComponentInstanceMap[key]
					if prevInstance != nil {
						for dKey := range prevInstance.DependencyKeys {
							affectedDepKeys[dKey] = true
						}
					}
				}
				{
					nextInstance := desiredState.ComponentInstanceMap[key]
					if nextInstance != nil {
						for dKey := range nextInstance.DependencyKeys {
							affectedDepKeys[dKey] = true
						}
					}
				}

				// if our dependency is affected, reset its deployed status to false (because actions are pending)
				for dKey := range affectedDepKeys {
					if _, ok := result.Status[dKey]; ok {
						result.Status[dKey].Deployed = false
					}
				}
			}
			return nil
		}),
		action.NewApplyResultUpdaterImpl(),
	)

}

func fetchReadinessStatusForDependencies(result *DependencyStatus, actualState *resolve.PolicyResolution, desiredState *resolve.PolicyResolution) {
	// TODO: this needs to be implemented. see https://github.com/Aptomi/aptomi/issues/315
}

func fetchEndpointsForDependencies(result *DependencyStatus, actualState *resolve.PolicyResolution) {
	for _, instance := range actualState.ComponentInstanceMap {
		for dKey := range instance.DependencyKeys {
			if _, ok := result.Status[dKey]; ok {
				if len(instance.Endpoints) > 0 {
					result.Status[dKey].Endpoints[instance.GetName()] = instance.Endpoints
				}
			}
		}
	}
}

/*
type dependencyResourcesWrapper struct {
	Resources plugin.Resources
}

func (g *dependencyResourcesWrapper) GetKind() string {
	return "dependencyResources"
}

func (api *coreAPI) handleDependencyResourcesGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := runtime.LastGen
	policy, _, err := api.store.GetPolicy(gen)
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	ns := params.ByName("ns")
	kind := lang.DependencyObject.Kind
	name := params.ByName("name")

	obj, err := policy.GetObject(kind, name, ns)
	if err != nil {
		panic(fmt.Sprintf("error while getting object %s/%s/%s in policy #%s", ns, kind, name, gen))
	}
	if obj == nil {
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusNotFound)
	}

	// once dependency is loaded, we need to find its state in the actual state
	dependency := obj.(*lang.Dependency)
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("Can't load actual state to get endpoints: %s", err))
	}

	plugins := api.pluginRegistryFactory()
	depKey := runtime.KeyForStorable(dependency)
	resources := make(plugin.Resources)
	for _, instance := range actualState.ComponentInstanceMap {
		if _, ok := instance.DependencyKeys[depKey]; ok {
			codePlugin, pluginErr := pluginForComponentInstance(instance, policy, plugins)
			if pluginErr != nil {
				panic(fmt.Sprintf("Can't get plugin for component instance %s: %s", instance.GetKey(), err))
			}
			if codePlugin == nil {
				continue
			}

			instanceResources, resErr := codePlugin.Resources(instance.GetDeployName(), instance.CalculatedCodeParams, event.NewLog(logrus.WarnLevel, "resources", false))
			if resErr != nil {
				panic(fmt.Sprintf("Error while getting deployment resources for component instance %s: %s", instance.GetKey(), resErr))
			}

			resources.Merge(instanceResources)
		}
	}

	api.contentType.WriteOne(writer, request, &dependencyResourcesWrapper{resources})
}

func pluginForComponentInstance(instance *resolve.ComponentInstance, policy *lang.Policy, plugins plugin.Registry) (plugin.CodePlugin, error) {
	serviceObj, err := policy.GetObject(lang.ServiceObject.Kind, instance.Metadata.Key.ServiceName, instance.Metadata.Key.Namespace)
	if err != nil {
		return nil, err
	}
	component := serviceObj.(*lang.Service).GetComponentsMap()[instance.Metadata.Key.ComponentName]

	if component == nil || component.Code == nil {
		return nil, nil
	}

	clusterName := instance.GetCluster()
	if len(clusterName) <= 0 {
		return nil, fmt.Errorf("component instance does not have cluster assigned: %v", instance.GetKey())
	}

	clusterObj, err := policy.GetObject(lang.ClusterObject.Kind, clusterName, runtime.SystemNS)
	if err != nil {
		return nil, err
	}
	if clusterObj == nil {
		return nil, fmt.Errorf("can't find cluster in policy: %s", clusterName)
	}
	cluster := clusterObj.(*lang.Cluster)

	return plugins.ForCodeType(cluster, component.Code.Type)
}
*/
