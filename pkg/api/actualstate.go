package api

import (
	"fmt"
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

func isDomainAdmin(user *lang.User, policy *lang.Policy) bool {
	systemNamespace := policy.Namespace[runtime.SystemNS]
	var aclResolver *lang.ACLResolver
	if systemNamespace != nil {
		aclResolver = lang.NewACLResolver(systemNamespace.ACLRules)
	} else {
		aclResolver = lang.NewACLResolver(make(map[string]*lang.Rule))
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

func (api *coreAPI) handleActualStateReset(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// Load current policy
	policy, genCurrent, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("error while loading current policy: %s", err))
	}

	// check that user is a domain admin
	user := api.getUserRequired(request)
	if !isDomainAdmin(user, policy) {
		panic(fmt.Sprintf("user is not allowed to perform actual state reset"))
	}

	// See if noop flag is set
	noop, noopErr := strconv.ParseBool(params.ByName("noop"))
	if noopErr != nil {
		noop = false
	}

	// If we are in noop mode
	if noop {
		// See that would happen if we reset the actual state, calculate and return resolution log + action plan
		eventLog := event.NewLog(logrus.InfoLevel, "api-state-reset-noop").AddConsoleHook(api.logLevel)
		desiredState := resolve.NewPolicyResolver(policy, api.externalData, eventLog).ResolveAllDependencies()
		actionPlan := diff.NewPolicyResolutionDiff(desiredState, resolve.NewPolicyResolution(true)).ActionPlan

		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: genCurrent,             // policy generation didn't change
			PolicyChanged:    false,                  // policy has not been updated in the store
			WaitForRevision:  runtime.MaxGeneration,  // nothing to wait for
			PlanAsText:       actionPlan.AsText(),    // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(), // return policy resolution log
		})

	} else {
		// Reset actual state in the store
		err := api.store.ResetActualState()
		if err != nil {
			panic(fmt.Sprintf("error while resetting actual state in the store: %s", err))
		}

		// Calculate and return resolution log + action plan
		eventLog := event.NewLog(logrus.InfoLevel, "api-state-reset").AddConsoleHook(api.logLevel)
		desiredState := resolve.NewPolicyResolver(policy, api.externalData, eventLog).ResolveAllDependencies()
		actionPlan := diff.NewPolicyResolutionDiff(desiredState, resolve.NewPolicyResolution(true)).ActionPlan

		// If there are changes, we need to wait for the next revision
		var waitForRevision = runtime.MaxGeneration
		if actionPlan.NumberOfActions() > 0 {
			revision, err := api.store.GetLastRevisionForPolicy(genCurrent)
			if err != nil {
				panic(fmt.Sprintf("error while loading last revision of the current policy: %s", err))
			}
			waitForRevision = revision.GetGeneration().Next()
		}

		api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
			TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
			PolicyGeneration: genCurrent,             // policy didn't change
			PolicyChanged:    false,                  // have any policy object in the store been changed or not
			WaitForRevision:  waitForRevision,        // which revision to wait for
			PlanAsText:       actionPlan.AsText(),    // return action plan, so it can be printed by the client
			EventLog:         eventLog.AsAPIEvents(), // return policy resolution log
		})

		// signal to the channel that actual state has changed, that will trigger the enforcement right away
		api.runDesiredStateEnforcement <- true
	}

}
