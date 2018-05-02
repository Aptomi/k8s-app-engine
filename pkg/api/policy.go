package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action/component"
	"github.com/Aptomi/aptomi/pkg/engine/diff"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/event"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

// PolicyUpdateResult represents results for the policy update request (estimated list of actions to be executed to
// update existing actual state to the desired state)
type PolicyUpdateResult struct {
	runtime.TypeKind `yaml:",inline"`
	PolicyGeneration runtime.Generation
	PolicyChanged    bool
	Actions          []string
}

// GetDefaultColumns returns default set of columns to be displayed
func (result *PolicyUpdateResult) GetDefaultColumns() []string {
	return []string{"Policy Changes", "Instance Changes"}
}

// AsColumns returns PolicyUpdateResult representation as columns
func (result *PolicyUpdateResult) AsColumns() map[string]string {
	var policyChangesStr string
	if result.PolicyChanged {
		policyChangesStr = fmt.Sprintf("Gen %d -> %d", result.PolicyGeneration-1, result.PolicyGeneration)
	} else {
		policyChangesStr = fmt.Sprintf("Gen %d (none)", result.PolicyGeneration)
	}
	var instanceChangesStr string
	filteredActions := filterImportantActionKeys(result.Actions)
	if len(filteredActions) > 0 {
		instanceChangesStr = strings.Join(filteredActions, "\n")
	} else {
		instanceChangesStr = "(none)"
	}
	return map[string]string{
		"Policy Changes":   policyChangesStr,
		"Instance Changes": instanceChangesStr,
	}
}

func filterImportantActionKeys(actions []string) []string {
	filtered := make([]string, 0)

	// remove #root, keep only create/update/delete actions
	importantActionKinds := map[string]string{
		component.CreateActionObject.Kind: "[+]",
		component.UpdateActionObject.Kind: "[*]",
		component.DeleteActionObject.Kind: "[-]",
	}

	for _, action := range actions {
		if strings.HasSuffix(action, "#root") {
			action = strings.TrimSuffix(action, "#root")
		}

		for kind, kindShort := range importantActionKinds {
			if strings.HasPrefix(action, kind+runtime.KeySeparator) {
				actionStr := kindShort + " " + strings.TrimPrefix(action, kind+runtime.KeySeparator)
				filtered = append(filtered, actionStr)
			}
		}
	}

	// sort
	sort.Strings(filtered)

	// shorten children
	result := make([]string, 0)
	for idx := 0; idx < len(filtered); {
		// skip empty entries, just in case
		if len(filtered[idx]) <= 0 {
			idx++
			continue
		}

		// move into result, if result is empty
		if len(result) == 0 {
			result = append(result, filtered[idx])
			idx++
			continue
		}

		// add directly if there is no need to shorten children
		if !strings.HasPrefix(filtered[idx], result[len(result)-1]) {
			result = append(result, filtered[idx])
			idx++
			continue
		}

		// process the block of strings/children and shorten them
		shorts := make([]string, 0)
		for idx < len(filtered) && strings.HasPrefix(filtered[idx], result[len(result)-1]) {
			// remove prefix
			short := strings.TrimPrefix(filtered[idx], result[len(result)-1])
			// remove #
			if strings.HasPrefix(short, "#") {
				short = strings.TrimPrefix(short, "#")
			}
			// pad it
			if len(short) > 0 {
				shorts = append(shorts, short)
			}
			idx++
		}

		result = append(result, "\t -> "+strings.Join(shorts, ", "))
	}

	return result
}

func (api *coreAPI) handlePolicyUpdate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	objects := api.readLang(request)

	user := api.getUserRequired(request)

	// Verify ACL for updated objects
	policy, _, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("Error while loading current policy: %s", err))
	}
	for _, obj := range objects {
		errAdd := policy.AddObject(obj)
		if errAdd != nil {
			panic(fmt.Sprintf("Error while adding updated object to policy: %s", errAdd))
		}
		errManage := policy.View(user).ManageObject(obj)
		if errManage != nil {
			panic(fmt.Sprintf("Error while adding updated object to policy: %s", errManage))
		}
	}

	err = policy.Validate()
	if err != nil {
		panic(fmt.Sprintf("Updated policy is invalid: %s", err))
	}

	// Validate clusters using corresponding cluster plugins if policy is valid
	plugins := api.pluginRegistryFactory()
	for _, obj := range objects {
		if cluster, ok := obj.(*lang.Cluster); ok {
			plugin, pluginErr := plugins.ForCluster(cluster)
			if pluginErr != nil {
				panic(fmt.Sprintf("Error while getting cluster plugin for cluster %s of type %s: %s", cluster.Name, cluster.Type, pluginErr))
			}

			valErr := plugin.Validate()
			if valErr != nil {
				panic(fmt.Sprintf("Error while validating cluster %s of type %s: %s", cluster.Name, cluster.Type, valErr))
			}
		}
	}

	changed, policyData, err := api.store.UpdatePolicy(objects, user.Name)
	if err != nil {
		panic(fmt.Sprintf("Error while updating policy: %s", err))
	}

	api.getPolicyUpdateResult(writer, request, changed, policyData)

	if changed {
		// signal to the channel that policy has changed, that will trigger the enforcement right away
		api.policyChanged <- true
	}
}

func (api *coreAPI) handlePolicyDelete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	objects := api.readLang(request)

	user := api.getUserRequired(request)

	// Verify ACL for updated objects
	currentPolicy, _, err := api.store.GetPolicy(runtime.LastGen)
	if err != nil {
		panic(fmt.Sprintf("Error while loading current policy: %s", err))
	}
	for _, obj := range objects {
		errManage := currentPolicy.View(user).ManageObject(obj)
		if errManage != nil {
			panic(fmt.Sprintf("Error while removing object from policy: %s", errManage))
		}
		currentPolicy.RemoveObject(obj)
	}

	err = currentPolicy.Validate()
	if err != nil {
		panic(fmt.Sprintf("Updated policy is invalid: %s", err))
	}

	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("Error while getting actual state: %s", err))
	}

	err = actualState.Validate(currentPolicy)
	if err != nil {
		panic(fmt.Sprintf("Updated policy is invalid: %s", err))
	}

	changed, policyData, err := api.store.DeleteFromPolicy(objects, user.Name)
	if err != nil {
		panic(fmt.Sprintf("Error while deleting from policy: %s", err))
	}

	api.getPolicyUpdateResult(writer, request, changed, policyData)

	if changed {
		// signal to the channel that policy has changed, that will trigger the enforcement right away
		api.policyChanged <- true
	}
}

func (api *coreAPI) getPolicyUpdateResult(writer http.ResponseWriter, request *http.Request, changed bool, policyData *engine.PolicyData) {
	desiredPolicyGen := policyData.GetGeneration()
	desiredPolicy, _, err := api.store.GetPolicy(desiredPolicyGen)
	if err != nil {
		panic(fmt.Sprintf("Error while getting desiredPolicy: %s", err))
	}
	if desiredPolicy == nil {
		panic(fmt.Sprintf("Can't read policy right after updating it"))
	}

	actualState, err := api.store.GetActualState()
	if err != nil {
		panic(fmt.Sprintf("Error while getting actual state: %s", err))
	}

	// todo we should resolve before saving policy => add Mutex for this method to make sure it's safe
	// todo: add request id to the event log scope
	eventLog := event.NewLog("api-policy-update", true)
	resolver := resolve.NewPolicyResolver(desiredPolicy, api.externalData, eventLog)
	desiredState := resolver.ResolveAllDependencies()
	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState)

	// TODO: we need to start showing dependency status in API result, as well as links/cmds to view logs

	actions := []string{}
	_ = stateDiff.ActionPlan.Apply(action.WrapSequential(func(act action.Base) error {
		actions = append(actions, act.GetName())
		return nil
	}))

	api.contentType.WriteOne(writer, request, &PolicyUpdateResult{
		TypeKind:         PolicyUpdateResultObject.GetTypeKind(),
		PolicyGeneration: desiredPolicyGen,
		PolicyChanged:    changed,
		Actions:          actions,
	})
}
