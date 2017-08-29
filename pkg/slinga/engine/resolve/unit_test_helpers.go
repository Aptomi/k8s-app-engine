package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/stretchr/testify/assert"
	"testing"
)

func loadPolicyAndResolve(t *testing.T) (*PolicyNamespace, *ServiceUsageState) {
	policy := LoadUnitTestsPolicy("../../testdata/unittests")
	return policy, resolvePolicy(t, policy)
}

func resolvePolicy(t *testing.T, policy *PolicyNamespace) *ServiceUsageState {
	userLoader := NewUserLoaderFromDir("../../testdata/unittests")
	return resolvePolicyInternal(t, policy, userLoader)
}

func resolvePolicyInternal(t *testing.T, policy *PolicyNamespace, userLoader UserLoader) *ServiceUsageState {
	resolver := NewPolicyResolver(policy, userLoader)
	result, err := resolver.ResolveAllDependencies()
	if !assert.Nil(t, err, "PolicyNamespace usage should be resolved without errors") {
		t.FailNow()
	}
	return result.State
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
