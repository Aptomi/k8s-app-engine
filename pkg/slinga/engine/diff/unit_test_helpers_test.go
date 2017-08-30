package diff

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func loadPolicyAndResolve(t *testing.T) *ResolvedState {
	policy := LoadUnitTestsPolicy("../../testdata/unittests")
	return resolvePolicy(t, policy)
}

func resolvePolicy(t *testing.T, policy *PolicyNamespace) *ResolvedState {
	userLoader := getUserLoader()
	return resolvePolicyInternal(t, policy, userLoader)
}
func getUserLoader() UserLoader {
	return NewUserLoaderFromDir("../../testdata/unittests")
}

func resolvePolicyInternal(t *testing.T, policy *PolicyNamespace, userLoader UserLoader) *ResolvedState {
	resolver := NewPolicyResolver(policy, userLoader)
	result, err := resolver.ResolveAllDependencies()
	if !assert.Nil(t, err, "PolicyNamespace usage should be resolved without errors") {
		t.FailNow()
	}
	return result
}

// TODO: this has to be changed to use the new serialization code instead of serializing to YAML
func emulateSaveAndLoadState(resolvedState *ResolvedState) *ResolvedState {
	// Save and load policy
	policyNew := PolicyNamespace{}
	yaml.DeserializeObject(yaml.SerializeObject(resolvedState.Policy), &policyNew)

	// Save and load state
	stateNew := ServiceUsageState{}
	yaml.DeserializeObject(yaml.SerializeObject(resolvedState.State), &stateNew)

	return NewResolvedState(&policyNew, &stateNew, resolvedState.UserLoader)
}

func getInstanceInternal(t *testing.T, key string, usageData *ServiceUsageData) *ComponentInstance {
	instance, ok := usageData.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance in usage data: "+key) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, serviceName string, contextName string, allocationKeysResolved []string, componentName string, policy *PolicyNamespace, state *ServiceUsageState) *ComponentInstance {
	key := NewComponentInstanceKey(serviceName, policy.Contexts[contextName], allocationKeysResolved, policy.Services[serviceName].GetComponentsMap()[componentName])
	return getInstanceInternal(t, key.GetKey(), state.ResolvedData)
}

func verifyDiff(t *testing.T, diff *ServiceUsageStateDiff, newRevision bool, componentInstantiate int, componentDestruct int, componentUpdate int, componentAttachDependency int, componentDetachDependency int) {
	assert.Equal(t, newRevision, diff.ShouldGenerateNewRevision(), "Diff: should generate new revision")
	assert.Equal(t, componentInstantiate, len(diff.ComponentInstantiate), "Diff: component instantiations")
	assert.Equal(t, componentDestruct, len(diff.ComponentDestruct), "Diff: component destructions")
	assert.Equal(t, componentUpdate, len(diff.ComponentUpdate), "Diff: component updates")
	assert.Equal(t, componentAttachDependency, len(diff.ComponentAttachDependency), "Diff: dependencies attached to components")
	assert.Equal(t, componentDetachDependency, len(diff.ComponentDetachDependency), "Diff: dependencies removed from components")
}

type componentTimes struct {
	timePrevCreated time.Time
	timePrevUpdated time.Time
	timeNextCreated time.Time
	timeNextUpdated time.Time
}

func getTimes(t *testing.T, key string, u1 *ServiceUsageState, u2 *ServiceUsageState) componentTimes {
	return componentTimes{
		timePrevCreated: getInstanceInternal(t, key, u1.ResolvedData).CreatedOn,
		timePrevUpdated: getInstanceInternal(t, key, u1.ResolvedData).UpdatedOn,
		timeNextCreated: getInstanceInternal(t, key, u2.ResolvedData).CreatedOn,
		timeNextUpdated: getInstanceInternal(t, key, u2.ResolvedData).UpdatedOn,
	}
}

func getTimesNext(t *testing.T, key string, u2 *ServiceUsageState) componentTimes {
	return componentTimes{
		timeNextCreated: getInstanceInternal(t, key, u2.ResolvedData).CreatedOn,
		timeNextUpdated: getInstanceInternal(t, key, u2.ResolvedData).UpdatedOn,
	}
}
