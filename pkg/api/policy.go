package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
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

func (api *coreAPI) handlePolicyUpdate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) { // nolint: gocyclo
	objects := api.readLang(request)
	user := api.getUserRequired(request)

	// Load the latest policy
	policy, policyGen, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// load the latest revision for the given policy
	revision, err := api.store.GetLastRevisionForPolicy(policyGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading latest revision from the store: %s", err))
	}

	// load desired state
	desiredState, err := api.store.GetDesiredState(revision, policy, api.externalData)
	if err != nil {
		panic(fmt.Sprintf("can't load desired state from revision: %s", err))
	}

	// Make a copy of the latest policy, so we can apply changes to it
	policyUpdated, _, err := api.store.GetPolicy(policyGen)
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

	// Process policy changes, calculate resolution log and action plan
	eventLog := event.NewLog(logLevel, "api-policy-update").AddConsoleHook(api.logLevel)
	desiredStateUpdated := resolve.NewPolicyResolver(policyUpdated, api.externalData, eventLog).ResolveAllDependencies()
	err = desiredStateUpdated.Validate(policyUpdated)
	if err != nil {
		panic(fmt.Sprintf("policy change cannon be made: %s", err))
	}

	actionPlan := diff.NewPolicyResolutionDiff(desiredStateUpdated, desiredState).ActionPlan

	// If we are in noop mode, just return expected changes in a form of an action plan
	if noop {
		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: policyGen,              // policy generation didn't change
			PolicyChanged:    false,                  // policy has not been updated in the store
			WaitForRevision:  runtime.MaxGeneration,  // nothing to wait for
			PlanAsText:       actionPlan.AsText(),    // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(), // return policy resolution log
		})
		return
	}

	// Here we need to take mutex to handle policy and revision updates
	api.policyAndRevisionUpdateMutex.Lock()
	defer api.policyAndRevisionUpdateMutex.Unlock()

	// Make object changes in the store
	changed, policyData, err := api.store.UpdatePolicy(objects, user.Name)
	if err != nil {
		panic(fmt.Sprintf("error while updating objects in policy: %s", err))
	}

	// If there are changes, create a new revision and say that we should wait for it
	waitForRevision := runtime.MaxGeneration
	if changed {
		newRevision, newRevisionErr := api.store.NewRevision(policyData.GetGeneration(), desiredStateUpdated, false)
		if newRevisionErr != nil {
			panic(fmt.Errorf("unable to create new revision for policy gen %d", policyGen))
		}
		waitForRevision = newRevision.GetGeneration()
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

func (api *coreAPI) handlePolicyDelete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	objects := api.readLang(request)
	user := api.getUserRequired(request)

	// Load the latest policy
	policy, policyGen, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// Load the latest revision for the given policy
	revision, err := api.store.GetLastRevisionForPolicy(policyGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading latest revision from the store: %s", err))
	}

	// Load desired state
	desiredState, err := api.store.GetDesiredState(revision, policy, api.externalData)
	if err != nil {
		panic(fmt.Sprintf("can't load desired state from revision: %s", err))
	}

	// Make a copy of the latest policy, so we can apply changes to it
	policyUpdated, _, err := api.store.GetPolicy(policyGen)
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

	// Process policy changes, calculate and return resolution log + action plan
	eventLog := event.NewLog(logLevel, "api-policy-delete").AddConsoleHook(api.logLevel)
	desiredStateUpdated := resolve.NewPolicyResolver(policyUpdated, api.externalData, eventLog).ResolveAllDependencies()
	err = desiredStateUpdated.Validate(policyUpdated)
	if err != nil {
		panic(fmt.Sprintf("policy change cannon be made: %s", err))
	}

	actionPlan := diff.NewPolicyResolutionDiff(desiredStateUpdated, desiredState).ActionPlan

	// If we are in noop mode, just return expected changes in a form of an action plan
	if noop {
		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: policyGen,              // policy generation didn't change
			PolicyChanged:    false,                  // policy has not been updated in the store
			WaitForRevision:  runtime.MaxGeneration,  // nothing to wait for
			PlanAsText:       actionPlan.AsText(),    // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(), // return policy resolution log
		})
		return
	}

	// Here we need to take mutex to handle policy and revision updates
	api.policyAndRevisionUpdateMutex.Lock()
	defer api.policyAndRevisionUpdateMutex.Unlock()

	// Make object changes in the store
	changed, policyData, err := api.store.DeleteFromPolicy(objects, user.Name)
	if err != nil {
		panic(fmt.Sprintf("error while deleting objects from policy: %s", err))
	}

	// If there are changes, create a new revision and say that we should wait for it
	waitForRevision := runtime.MaxGeneration
	if changed {
		newRevision, newRevisionErr := api.store.NewRevision(policyData.GetGeneration(), desiredStateUpdated, false)
		if newRevisionErr != nil {
			panic(fmt.Errorf("unable to create new revision for policy gen %d", policyGen))
		}
		waitForRevision = newRevision.GetGeneration()
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
