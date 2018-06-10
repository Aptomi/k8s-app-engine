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
	"strconv"
)

func (api *coreAPI) handlePolicyGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := params.ByName("gen")

	if len(gen) == 0 {
		gen = strconv.Itoa(int(runtime.LastGen))
	}

	policyData, err := api.store.GetPolicyData(runtime.ParseGeneration(gen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	if policyData == nil {
		// policy with the given generation not found
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusNotFound)
	} else {
		api.contentType.WriteOne(writer, request, policyData)
	}
}

func (api *coreAPI) handlePolicyObjectGet(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	gen := params.ByName("gen")

	if len(gen) == 0 {
		gen = strconv.Itoa(int(runtime.LastGen))
	}

	policy, _, err := api.store.GetPolicy(runtime.ParseGeneration(gen))
	if err != nil {
		panic(fmt.Sprintf("error while getting requested policy: %s", err))
	}

	ns := params.ByName("ns")
	kind := params.ByName("kind")
	name := params.ByName("name")

	obj, err := policy.GetObject(kind, name, ns)
	if err != nil {
		panic(fmt.Sprintf("error while getting object %s/%s/%s in policy #%s", ns, kind, name, gen))
	}
	if obj == nil {
		api.contentType.WriteOneWithStatus(writer, request, nil, http.StatusNotFound)
	}

	api.contentType.WriteOne(writer, request, obj)
}

// PolicyUpdateResultObject is an informational data structure with Kind and Constructor for PolicyUpdateResult
var PolicyUpdateResultObject = &runtime.Info{
	Kind:        "policy-update-result",
	Constructor: func() runtime.Object { return &PolicyUpdateResult{} },
}

// PolicyUpdateResult represents results of the policy update request, including action plan and event log
type PolicyUpdateResult struct {
	runtime.TypeKind `yaml:",inline"`
	PolicyGeneration runtime.Generation
	PolicyChanged    bool
	WaitForRevision  runtime.Generation
	PlanAsText       *action.PlanAsText
	EventLog         []*event.APIEvent
}

// GetDefaultColumns returns default set of columns to be displayed
func (result *PolicyUpdateResult) GetDefaultColumns() []string {
	return []string{"Policy Generation", "Action Plan"}
}

// AsColumns returns PolicyUpdateResult representation as columns
func (result *PolicyUpdateResult) AsColumns() map[string]string {
	var policyChangesStr string
	if result.PolicyChanged {
		policyChangesStr = fmt.Sprintf("%d -> %d", result.PolicyGeneration-1, result.PolicyGeneration)
	} else {
		policyChangesStr = fmt.Sprintf("%d", result.PolicyGeneration)
	}
	var actionPlanStr = result.PlanAsText.String()
	if len(actionPlanStr) <= 0 {
		actionPlanStr = "(none)"
	}
	return map[string]string{
		"Policy Generation": policyChangesStr,
		"Action Plan":       actionPlanStr,
	}
}

func (api *coreAPI) handlePolicyUpdate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	objects := api.readLang(request)
	user := api.getUserRequired(request)

	// Load current policy
	policyUpdated, genCurrent, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// Store copy of the current policy before we modify it
	policy, _, err := api.store.GetPolicy(genCurrent)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// Verify that user has permissions to create and update objects
	for _, obj := range objects {
		errAdd := policyUpdated.AddObject(obj)
		if errAdd != nil {
			panic(fmt.Sprintf("error while adding updated object to policy: %s", errAdd))
		}
		errManage := policyUpdated.View(user).ManageObject(obj)
		if errManage != nil {
			panic(fmt.Sprintf("error while adding updated object to policy: %s", errManage))
		}
	}

	// Check that the policy is valid
	err = policyUpdated.Validate()
	if err != nil {
		panic(fmt.Sprintf("updated policy is invalid: %s", err))
	}

	// Validate clusters using corresponding cluster plugins and make sure there are no conflicts
	plugins := api.pluginRegistryFactory()
	for _, obj := range objects {
		// if a cluster was supplied, then
		if cluster, ok := obj.(*lang.Cluster); ok {
			// if a cluster is already present in the policy, tell a user that it can't be modified
			objExisting, _ := policy.GetObject(lang.ClusterObject.Kind, cluster.Name, cluster.Namespace)
			if objExisting != nil {
				panic(fmt.Sprintf("modification of existing cluster objects is not allowed: %s needs to be deleted first", cluster.Name))
			}

			// validate via plugin that connection to it can be established
			plugin, pluginErr := plugins.ForCluster(cluster)
			if pluginErr != nil {
				panic(fmt.Sprintf("error while getting cluster plugin for cluster %s of type %s: %s", cluster.Name, cluster.Type, pluginErr))
			}

			valErr := plugin.Validate()
			if valErr != nil {
				panic(fmt.Sprintf("error while validating cluster %s of type %s: %s", cluster.Name, cluster.Type, valErr))
			}
		}
	}

	// See if noop flag is set
	noop, noopErr := strconv.ParseBool(params.ByName("noop"))
	if noopErr != nil {
		noop = false
	}

	// See what log level is set
	logLevel, logLevelErr := logrus.ParseLevel(params.ByName("loglevel"))
	if logLevelErr != nil {
		logLevel = logrus.WarnLevel
	}

	// If we are in noop mode
	if noop {
		// Process policy changes, calculate and return resolution log + action plan
		eventLog := event.NewLog(logLevel, "api-policy-update-noop").AddConsoleHook(api.logLevel)
		desiredStatePrev := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "prev")).ResolveAllDependencies()
		desiredState := resolve.NewPolicyResolver(policyUpdated, api.externalData, eventLog).ResolveAllDependencies()
		actionPlan := diff.NewPolicyResolutionDiff(desiredState, desiredStatePrev).ActionPlan

		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: genCurrent,             // policy generation didn't change
			PolicyChanged:    false,                  // policy has not been updated in the store
			WaitForRevision:  runtime.MaxGeneration,  // nothing to wait for
			PlanAsText:       actionPlan.AsText(),    // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(), // return policy resolution log
		})

	} else {
		// Make object changes in the store
		changed, policyData, err := api.store.UpdatePolicy(objects, user.Name)
		if err != nil {
			panic(fmt.Sprintf("error while updating objects in policy: %s", err))
		}

		// Process policy changes, calculate and return resolution log + action plan
		eventLog := event.NewLog(logLevel, "api-policy-update").AddConsoleHook(api.logLevel)
		desiredStatePrev := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "prev")).ResolveAllDependencies()
		desiredState := resolve.NewPolicyResolver(policyUpdated, api.externalData, eventLog).ResolveAllDependencies()
		actionPlan := diff.NewPolicyResolutionDiff(desiredState, desiredStatePrev).ActionPlan

		// If there are changes, we need to wait for the next revision
		var waitForRevision = runtime.MaxGeneration
		if changed {
			revision, err := api.store.GetLastRevisionForPolicy(genCurrent)
			if err != nil {
				panic(fmt.Sprintf("error while loading last revision of the current policy: %s", err))
			}
			waitForRevision = revision.GetGeneration().Next()
		}

		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: policyData.GetGeneration(), // policy now has a new generation
			PolicyChanged:    changed,                    // have any policy object in the store been changed or not
			WaitForRevision:  waitForRevision,            // which revision to wait for
			PlanAsText:       actionPlan.AsText(),        // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(),     // return policy resolution log
		})

		if changed {
			// signal to the channel that policy has changed, that will trigger the enforcement right away
			api.runDesiredStateEnforcement <- true
		}
	}
}

func (api *coreAPI) handlePolicyDelete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	objects := api.readLang(request)
	user := api.getUserRequired(request)

	// Load current policy
	policyUpdated, genCurrent, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// Store copy of the current policy before we modify it
	policy, _, err := api.store.GetPolicy(genCurrent)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// Verify that user has permissions to delete objects
	for _, obj := range objects {
		errManage := policyUpdated.View(user).ManageObject(obj)
		if errManage != nil {
			panic(fmt.Sprintf("Error while removing object from policy: %s", errManage))
		}
		policyUpdated.RemoveObject(obj)
	}

	err = policyUpdated.Validate()
	if err != nil {
		panic(fmt.Sprintf("Updated policy is invalid: %s", err))
	}

	desiredStateTmp := resolve.NewPolicyResolver(policyUpdated, api.externalData, event.NewLog(logrus.WarnLevel, "tmp")).ResolveAllDependencies()
	err = desiredStateTmp.Validate(policyUpdated)
	if err != nil {
		panic(fmt.Sprintf("Updated policy is invalid: %s", err))
	}

	// See if noop flag is set
	noop, noopErr := strconv.ParseBool(params.ByName("noop"))
	if noopErr != nil {
		noop = false
	}

	// See what log level is set
	logLevel, logLevelErr := logrus.ParseLevel(params.ByName("loglevel"))
	if logLevelErr != nil {
		logLevel = logrus.WarnLevel
	}

	// If we are in noop mode
	if noop {
		// Process policy changes, calculate and return resolution log + action plan
		eventLog := event.NewLog(logLevel, "api-policy-delete-noop").AddConsoleHook(api.logLevel)
		desiredStatePrev := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "prev")).ResolveAllDependencies()
		desiredState := resolve.NewPolicyResolver(policyUpdated, api.externalData, eventLog).ResolveAllDependencies()
		actionPlan := diff.NewPolicyResolutionDiff(desiredState, desiredStatePrev).ActionPlan

		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: genCurrent,             // policy generation didn't change
			PolicyChanged:    false,                  // policy has not been updated in the store
			WaitForRevision:  runtime.MaxGeneration,  // nothing to wait for
			PlanAsText:       actionPlan.AsText(),    // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(), // return policy resolution log
		})

	} else {
		// Make object changes in the store
		changed, policyData, err := api.store.DeleteFromPolicy(objects, user.Name)
		if err != nil {
			panic(fmt.Sprintf("error while deleting objects from policy: %s", err))
		}

		// Process policy changes, calculate and return resolution log + action plan
		eventLog := event.NewLog(logLevel, "api-policy-delete").AddConsoleHook(api.logLevel)
		desiredStatePrev := resolve.NewPolicyResolver(policy, api.externalData, event.NewLog(logrus.WarnLevel, "prev")).ResolveAllDependencies()
		desiredState := resolve.NewPolicyResolver(policyUpdated, api.externalData, eventLog).ResolveAllDependencies()
		actionPlan := diff.NewPolicyResolutionDiff(desiredState, desiredStatePrev).ActionPlan

		// If there are changes, we need to wait for the next revision
		var waitForRevision = runtime.MaxGeneration
		if changed {
			revision, err := api.store.GetLastRevisionForPolicy(genCurrent)
			if err != nil {
				panic(fmt.Sprintf("error while loading last revision of the current policy: %s", err))
			}
			waitForRevision = revision.GetGeneration().Next()
		}

		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: policyData.GetGeneration(), // policy now has a new generation
			PolicyChanged:    changed,                    // have any policy object in the store been changed or not
			WaitForRevision:  waitForRevision,            // which revision to wait for
			PlanAsText:       actionPlan.AsText(),        // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(),     // return policy resolution log
		})

		if changed {
			// signal to the channel that policy has changed, that will trigger the enforcement right away
			api.runDesiredStateEnforcement <- true
		}
	}

}
