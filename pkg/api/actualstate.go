package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func isDomainAdmin(user *lang.User, policy *lang.Policy) bool {
	systemNamespace := policy.Namespace[runtime.SystemNS]
	var aclResolver *lang.ACLResolver
	if systemNamespace != nil {
		aclResolver = lang.NewACLResolver(systemNamespace.ACLRules)
	} else {
		aclResolver = lang.NewACLResolver(make(map[string]*lang.ACLRule))
	}

	roleMap, errRoleMap := aclResolver.GetUserRoleMap(user)
	if errRoleMap != nil {
		panic(fmt.Sprintf("error while getting user role map: %s", errRoleMap))
	}

	if _, ok := roleMap[lang.DomainAdmin.ID]; ok {
		return true
	}

	return false
}

func (api *coreAPI) handleStateEnforce(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// Load current policy
	policy, policyGen, err := api.registry.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading latest policy: %s", err))
	}

	// check that user is a domain admin
	user := api.getUserRequired(request)
	if !isDomainAdmin(user, policy) {
		panic(fmt.Sprintf("user is not allowed to trigger actual state enforcement"))
	}

	// See if noop flag is set
	noop, noopErr := strconv.ParseBool(params.ByName("noop"))
	if noopErr != nil {
		noop = false
	}

	// See that would happen if we reset the actual state, calculate resolution log and action plan
	resolveLog := event.NewLog(logrus.InfoLevel, "api-state-enforce").AddConsoleHook(api.logLevel)
	desiredState := resolve.NewPolicyResolver(policy, api.externalData, resolveLog).ResolveAllClaims()
	actionPlan := diff.NewPolicyResolutionDiff(desiredState, resolve.NewPolicyResolution()).ActionPlan

	// If we are in noop mode, just return expected changes in a form of an action plan
	if noop {
		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: policyGen,                // policy generation didn't change
			PolicyChanged:    false,                    // policy has not been updated in the registry
			WaitForRevision:  runtime.MaxGeneration,    // nothing to wait for
			PlanAsText:       actionPlan.AsText(),      // return action plan, so it can be printed by the client
			EventLog:         resolveLog.AsAPIEvents(), // return policy resolution log
		})
		return
	}

	// Keep policy the same, but create another special revision for it to enforce the state
	revisionGen := api.createStateEnforceRevision(policyGen, desiredState, actionPlan)

	api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
		TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
		PolicyGeneration: policyGen,                // policy didn't change
		PolicyChanged:    false,                    // have any policy object in the registry been changed or not
		WaitForRevision:  revisionGen,              // which revision to wait for
		PlanAsText:       actionPlan.AsText(),      // return action plan, so it can be printed by the client
		EventLog:         resolveLog.AsAPIEvents(), // return policy resolution log
	})

	// signal to the channel that actual state has changed, that will trigger the enforcement right away
	api.runDesiredStateEnforcement <- true
}

func (api *coreAPI) createStateEnforceRevision(policyGen runtime.Generation, desiredState *resolve.PolicyResolution, actionPlan *action.Plan) runtime.Generation {
	// Here we need to take mutex to handle policy and revision updates
	api.policyAndRevisionUpdateMutex.Lock()
	defer api.policyAndRevisionUpdateMutex.Unlock()

	// If there are changes, create new special revision for enforcing state and say that we should wait for it
	var revisionGen = runtime.MaxGeneration
	if actionPlan.NumberOfActions() > 0 {
		// If there are changes, create a new revision and say that we should wait for it
		newRevision, newRevisionErr := api.registry.NewRevision(policyGen, desiredState, true)
		if newRevisionErr != nil {
			panic(fmt.Errorf("unable to create new revision for policy gen %d", policyGen))
		}
		revisionGen = newRevision.GetGeneration()
	}

	return revisionGen
}
