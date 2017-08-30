package resolve

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func loadUnitTestsPolicy() *PolicyNamespace {
	return LoadUnitTestsPolicy("../../testdata/unittests")
}

func loadPolicyAndResolve(t *testing.T) (*PolicyNamespace, *ServiceUsageState) {
	policy := loadUnitTestsPolicy()
	return policy, resolvePolicy(t, policy, ResSuccess)
}

func resolvePolicy(t *testing.T, policy *PolicyNamespace, expectedResult int) *ServiceUsageState {
	userLoader := NewUserLoaderFromDir("../../testdata/unittests")
	resolver := NewPolicyResolver(policy, userLoader)
	result, err := resolver.ResolveAllDependencies()
	if !assert.Equal(t, expectedResult != ResError, err == nil, "Policy resolution status (success vs. error)") || expectedResult == ResError {
		return nil
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
