package api

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/plugin"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// ClaimQueryFlag determines whether to query just claim deployment status, or both deployment + readiness/health checks status
type ClaimQueryFlag string

const (
	// ClaimQueryDeploymentStatusOnly prescribes only to query claim deployment status (i.e. actual state = desired state)
	ClaimQueryDeploymentStatusOnly ClaimQueryFlag = "deployed"

	// ClaimQueryDeploymentStatusAndReadiness prescribes to query both claim deployment status (i.e. actual state = desired state), as well as readiness status (i.e. health checks = passing)
	ClaimQueryDeploymentStatusAndReadiness ClaimQueryFlag = "ready"
)

// ClaimsStatusType is an informational data structure with Kind and Constructor for ClaimsStatus
var ClaimsStatusType = &runtime.TypeInfo{
	Kind:        "claims-status",
	Constructor: func() runtime.Object { return &ClaimsStatus{} },
}

// ClaimsStatus is a struct which holds status information for a set of given claims
type ClaimsStatus struct {
	runtime.TypeKind `yaml:",inline"`

	// map containing status by claim
	Status map[string]*ClaimStatus
}

// ClaimStatus is a struct which holds status information for an individual claim
type ClaimStatus struct {
	Found     bool
	Deployed  bool
	Ready     bool
	Endpoints map[string]map[string]string
}

func (api *coreAPI) handleClaimStatusGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// parse query mode flag (deployment status vs. readiness status) as well as the list of claim IDs
	flag := ClaimQueryFlag(params.ByName("queryFlag"))
	claimIds := strings.Split(params.ByName("idList"), ",")

	// load the latest policy
	policy, policyGen, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading latest policy from the store: %s", err))
	}

	// load the latest revision for the given policy
	revision, err := api.store.GetLastRevisionForPolicy(policyGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading latest revision from the store: %s", err))
	}

	// load desired state
	desiredState, err := api.store.GetDesiredState(revision)
	if err != nil {
		panic(fmt.Sprintf("can't load desired state from revision: %s", err))
	}

	// load actual state
	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("can't load actual state from the store: %s", err))
	}

	// initialize result
	result := &ClaimsStatus{
		TypeKind: ClaimsStatusType.GetTypeKind(),
		Status:   make(map[string]*ClaimStatus),
	}
	for _, claimID := range claimIds {
		parts := strings.Split(claimID, "^")
		cObj, err := policy.GetObject(lang.ClaimType.Kind, parts[1], parts[0])
		if cObj == nil || err != nil {
			claimKey := runtime.KeyFromParts(parts[0], lang.ClaimType.Kind, parts[1])
			result.Status[claimKey] = &ClaimStatus{
				Found:     false,
				Deployed:  false,
				Ready:     false,
				Endpoints: make(map[string]map[string]string),
			}
			continue
		}

		claim := cObj.(*lang.Claim) // nolint: errcheck
		resolved := desiredState.GetClaimResolution(claim).Resolved
		result.Status[runtime.KeyForStorable(claim)] = &ClaimStatus{
			Found:     true,
			Deployed:  resolved,
			Ready:     resolved,
			Endpoints: make(map[string]map[string]string),
		}
	}

	// fetch deployment status for claims
	fetchDeploymentStatusForClaims(result, actualState, desiredState)

	// fetch readiness status for claims, if we were asked to do so
	if flag == ClaimQueryDeploymentStatusAndReadiness {
		plugins := api.pluginRegistryFactory()
		fetchReadinessStatusForClaims(result, plugins, policy, actualState)
	}

	// fetch endpoints for claims
	fetchEndpointsForClaims(result, actualState)

	// return the result back
	api.contentType.WriteOne(writer, request, result)
}

func fetchDeploymentStatusForClaims(result *ClaimsStatus, actualState *resolve.PolicyResolution, desiredState *resolve.PolicyResolution) {
	// compare desired vs. actual state and see what's the claim status for every provided claim ID
	actionPlan := diff.NewPolicyResolutionDiff(desiredState, actualState).ActionPlan
	actionPlan.Apply(
		action.WrapSequential(func(act action.Interface) error {
			// if it's attach action is pending on component, let's see which particular claim it affects
			if dAction, ok := act.(*component.AttachClaimAction); ok {
				// reset status of this particular claim to false
				if _, affected := result.Status[dAction.ClaimKey]; affected {
					result.Status[dAction.ClaimKey].Deployed = false
					return nil
				}
			}

			// if it's detach action is pending on component, let's see which particular claim it affects
			if dAction, ok := act.(*component.DetachClaimAction); ok {
				// reset status of this particular claim to false
				if _, affected := result.Status[dAction.ClaimKey]; affected {
					result.Status[dAction.ClaimKey].Deployed = false
					return nil
				}
			}

			key, ok := act.DescribeChanges()["key"].(string)
			if ok && len(key) > 0 {
				// we found a component in diff, which is affected by the action. let's see if any of the claims are affected
				affectedClaimKeys := make(map[string]bool)
				{
					prevInstance := actualState.ComponentInstanceMap[key]
					if prevInstance != nil {
						for claimKey := range prevInstance.ClaimKeys {
							affectedClaimKeys[claimKey] = true
						}
					}
				}
				{
					nextInstance := desiredState.ComponentInstanceMap[key]
					if nextInstance != nil {
						for claimKey := range nextInstance.ClaimKeys {
							affectedClaimKeys[claimKey] = true
						}
					}
				}

				// if our claim is affected, reset its deployed status to false (because actions are pending)
				for claimKey := range affectedClaimKeys {
					if _, ok := result.Status[claimKey]; ok {
						result.Status[claimKey].Deployed = false
					}
				}
			}
			return nil
		}),
		action.NewApplyResultUpdaterImpl(),
	)

	for _, instance := range actualState.ComponentInstanceMap {
		if instance.IsCode && !instance.EndpointsUpToDate {
			for claimKey := range instance.ClaimKeys {
				if _, ok := result.Status[claimKey]; ok {
					result.Status[claimKey].Deployed = false
				}
			}
		}
	}
}

func fetchReadinessStatusForClaims(result *ClaimsStatus, plugins plugin.Registry, policy *lang.Policy, actualState *resolve.PolicyResolution) {
	// if claim is not deployed, it means it's not ready
	for claimKey := range result.Status {
		result.Status[claimKey].Ready = result.Status[claimKey].Ready && result.Status[claimKey].Deployed
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

		// we only need to query status of this component, if at least one claim is still Ready
		foundClaimsToCheck := false
		for claimKey := range instance.ClaimKeys {
			if _, ok := result.Status[claimKey]; ok && result.Status[claimKey].Ready {
				foundClaimsToCheck = true
				break
			}
		}

		// if we don't need to query status of this component instance, let's just return
		if !foundClaimsToCheck {
			continue
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

			// update status of claims
			dUpdateMutex.Lock()
			defer dUpdateMutex.Unlock()
			for claimKey := range instance.ClaimKeys {
				if _, ok := result.Status[claimKey]; ok {
					result.Status[claimKey].Ready = result.Status[claimKey].Ready && instanceStatus
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

func fetchEndpointsForClaims(result *ClaimsStatus, actualState *resolve.PolicyResolution) {
	for _, instance := range actualState.ComponentInstanceMap {
		for claimKey := range instance.ClaimKeys {
			if _, ok := result.Status[claimKey]; ok {
				if len(instance.Endpoints) > 0 {
					result.Status[claimKey].Endpoints[instance.GetName()] = instance.Endpoints
				}
			}
		}
	}
}
