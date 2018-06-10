package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
)

// DependencyQueryFlag determines whether to query just dependency deployment status, or both deployment + readiness/health checks status
type DependencyQueryFlag string

const (
	// DependencyQueryDeploymentStatusOnly prescribes only to query dependency deployment status (i.e. actual state = desired state)
	DependencyQueryDeploymentStatusOnly DependencyQueryFlag = "deployed"

	// DependencyQueryDeploymentStatusAndReadiness prescribes to query both dependency deployment status (i.e. actual state = desired state), as well as readiness status (i.e. health checks = passing)
	DependencyQueryDeploymentStatusAndReadiness DependencyQueryFlag = "ready"
)

// DependenciesStatusObject is an informational data structure with Kind and Constructor for DependenciesStatus
var DependenciesStatusObject = &runtime.Info{
	Kind:        "dependencies-status",
	Constructor: func() runtime.Object { return &DependenciesStatus{} },
}

// DependenciesStatus is a struct which holds status information for a set of given dependencies
type DependenciesStatus struct {
	runtime.TypeKind `yaml:",inline"`

	// map containing status by dependency
	Status map[string]*DependencyStatus
}

// DependencyStatus is a struct which holds status information for an individual dependency
type DependencyStatus struct {
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
	result := &DependenciesStatus{
		TypeKind: DependenciesStatusObject.GetTypeKind(),
		Status:   make(map[string]*DependencyStatus),
	}
	for _, depID := range dependencyIds {
		parts := strings.Split(depID, "^")
		dObj, err := policy.GetObject(lang.DependencyObject.Kind, parts[1], parts[0])
		if dObj == nil || err != nil {
			dKey := runtime.KeyFromParts(parts[0], lang.DependencyObject.Kind, parts[1])
			result.Status[dKey] = &DependencyStatus{
				Found:     false,
				Deployed:  false,
				Ready:     false,
				Endpoints: make(map[string]map[string]string),
			}
			continue
		}

		d := dObj.(*lang.Dependency)
		result.Status[runtime.KeyForStorable(d)] = &DependencyStatus{
			Found:     true,
			Deployed:  true,
			Ready:     true,
			Endpoints: make(map[string]map[string]string),
		}
	}

	// load actual and desired states
	desiredState := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "api-dependencies-status")).ResolveAllDependencies()
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("can't load actual state from the store: %s", err))
	}

	// fetch deployment status for dependencies
	fetchDeploymentStatusForDependencies(result, actualState, desiredState)

	// fetch readiness status for dependencies, if we were asked to do so
	if flag == DependencyQueryDeploymentStatusAndReadiness {
		plugins := api.pluginRegistryFactory()
		fetchReadinessStatusForDependencies(result, plugins, policy, actualState)
	}

	// fetch endpoints for dependencies
	fetchEndpointsForDependencies(result, actualState)

	// return the result back
	api.contentType.WriteOne(writer, request, result)
}

func fetchDeploymentStatusForDependencies(result *DependenciesStatus, actualState *resolve.PolicyResolution, desiredState *resolve.PolicyResolution) {
	// compare desired vs. actual state and see what's the dependency status for every provided dependency ID
	diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan.Apply(
		action.WrapSequential(func(act action.Base) error {
			// if it's attach action is pending on component, let's see which particular dependency it affects
			if dAction, ok := act.(*component.AttachDependencyAction); ok {
				// reset status of this particular dependency to false
				if _, affected := result.Status[dAction.DependencyID]; affected {
					result.Status[dAction.DependencyID].Deployed = false
					return nil
				}
			}

			// if it's detach action is pending on component, let's see which particular dependency it affects
			if dAction, ok := act.(*component.DetachDependencyAction); ok {
				// reset status of this particular dependency to false
				if _, affected := result.Status[dAction.DependencyID]; affected {
					result.Status[dAction.DependencyID].Deployed = false
					return nil
				}
			}

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

	for _, instance := range actualState.ComponentInstanceMap {
		if instance.IsCode && !instance.EndpointsUpToDate {
			for dKey := range instance.DependencyKeys {
				if _, ok := result.Status[dKey]; ok {
					result.Status[dKey].Deployed = false
				}
			}
		}
	}
}

func fetchReadinessStatusForDependencies(result *DependenciesStatus, plugins plugin.Registry, policy *lang.Policy, actualState *resolve.PolicyResolution) {
	// if dependency is not deployed, it means it's not ready
	for dKey := range result.Status {
		result.Status[dKey].Ready = result.Status[dKey].Ready && result.Status[dKey].Deployed
	}

	// update readiness
	dUpdateMutex := sync.Mutex{}
	var wg sync.WaitGroup
	errors := make(chan error, 1)
	for _, instance := range actualState.ComponentInstanceMap {
		// if component instance is not code, skip it
		if !instance.IsCode {
			continue
		}

		// we only need to query status of this component, if at least one dependency is still Ready
		foundDependenciesToCheck := false
		for dKey := range instance.DependencyKeys {
			if _, ok := result.Status[dKey]; ok && result.Status[dKey].Ready {
				foundDependenciesToCheck = true
				break
			}
		}

		// if we don't need to query status of this component instance, let's just return
		if !foundDependenciesToCheck {
			return
		}

		wg.Add(1)
		go func(instance *resolve.ComponentInstance) {
			// make sure we are converting panics into errors
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					select {
					case errors <- fmt.Errorf("panic: %s\n%s", err, string(debug.Stack())):
						// message sent
					default:
						// error was already there before, do nothing (but we have to keep an empty default block)
					}
				}
			}()

			// query status of this component instance
			codePlugin, err := pluginForComponentInstance(instance, policy, plugins)
			if err != nil {
				panic(fmt.Sprintf("Can't get plugin for component instance %s: %s", instance.GetKey(), err))
			}
			if codePlugin == nil {
				return
			}

			instanceStatus, err := codePlugin.Status(
				&plugin.CodePluginInvocationParams{
					DeployName:   instance.GetDeployName(),
					Params:       instance.CalculatedCodeParams,
					PluginParams: map[string]string{plugin.ParamTargetSuffix: instance.Metadata.Key.TargetSuffix},
					EventLog:     event.NewLog(logrus.WarnLevel, "resources-status"),
				},
			)
			if err != nil {
				panic(fmt.Sprintf("Error while getting deployment resources status for component instance %s: %s", instance.GetKey(), err))
			}

			// update status of dependencies
			dUpdateMutex.Lock()
			defer dUpdateMutex.Unlock()
			for dKey := range instance.DependencyKeys {
				if _, ok := result.Status[dKey]; ok {
					result.Status[dKey].Ready = result.Status[dKey].Ready && instanceStatus
				}
			}
		}(instance)
	}

	// wait until all go routines are over
	wg.Wait()

	// see if there were any errors
	select {
	case err := <-errors:
		panic(err)
	default:
		// no error, do nothing (but we have to keep an empty default block)
	}
}

func fetchEndpointsForDependencies(result *DependenciesStatus, actualState *resolve.PolicyResolution) {
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
