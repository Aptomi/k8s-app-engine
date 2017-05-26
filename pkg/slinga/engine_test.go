package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEngine(t *testing.T) {
	policy := LoadPolicyFromDir("testdata/unittests")
	users := LoadUsersFromDir("testdata/unittests")
	dependencies := LoadDependenciesFromDir("testdata/unittests")

	usageState := NewServiceUsageState(&policy, &dependencies)
	err := usageState.ResolveUsage(&users)

	// Check that policy resolution finished correctly
	assert.Equal(t, nil, err, "Policy usage should be resolved without errors")
	assert.Equal(t, 14, len(usageState.ResolvedLinks), "Policy resolution should result in correct amount of usage entries")

	// Check that parameter evaluates correctly
	v := usageState.ResolvedLinks["kafka#test#test-platform_services#component2"]
	paramsMap, ok := v.CalculatedCodeParams.(map[string]interface{})
	assert.Equal(t, true, ok, "Calculated Code Params should be map")
	assert.Equal(t, "zookeeper-test-test-platform-services-component2", paramsMap["address"], "Code parameter should be calculated correctly")
}

func TestServiceComponentsTopologicalOrder(t *testing.T) {
	state := LoadPolicyFromDir("testdata/unittests")
	service := state.Services["kafka"]

	c, err := service.getComponentsSortedTopologically()
	assert.Equal(t, nil, err, "Service components should be topologically sorted without errors")

	assert.Equal(t, len(c), 3, "Component topological sort should produce correct number of values")
	assert.Equal(t, "component1", c[0].Name, "Component topological sort should produce correct order")
	assert.Equal(t, "component2", c[1].Name, "Component topological sort should produce correct order")
	assert.Equal(t, "component3", c[2].Name, "Component topological sort should produce correct order")
}
