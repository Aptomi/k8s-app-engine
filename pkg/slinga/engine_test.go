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
	assert.Nil(t, err, "Policy usage should be resolved without errors")

	kafkaTest := usageState.ResolvedLinks["kafka#test#test-platform_services#component2"]
	kafkaProd := usageState.ResolvedLinks["kafka#prod#test-platform_services#component2"]
	assert.Equal(t, 1, len(kafkaTest.UserIds), "Only one user should have access to test")
	assert.Equal(t, "1", kafkaTest.UserIds[0], "Only Alice should have access to test")

	assert.Equal(t, 1, len(kafkaProd.UserIds), "Only one user should have access to prod")
	assert.Equal(t, "2", kafkaProd.UserIds[0], "Only Bob should have access to prod (Carol is compromised)")

	// Check that code parameters evaluate correctly
	paramsMap, ok := kafkaTest.CalculatedCodeParams.(map[interface{}]interface{})
	assert.Equal(t, true, ok, "Calculated Code Params should be map")
	assert.Equal(t, "zookeeper-test-test-platform-services-component2", paramsMap["address"], "Code parameter should be calculated correctly")

	// Check that discovery parameters evaluate correctly
	discoveryMap, ok := kafkaTest.CalculatedDiscovery.(map[interface{}]interface{})
	assert.Equal(t, true, ok, "Calculated Discovery should be map")
	assert.Equal(t, "kafka-kafka-test-test-platform-services-component2-url", discoveryMap["url"], "Discovery parameter should be calculated correctly")
}

func TestServiceComponentsTopologicalOrder(t *testing.T) {
	state := LoadPolicyFromDir("testdata/unittests")
	service := state.Services["kafka"]

	c, err := service.getComponentsSortedTopologically()
	assert.Nil(t, err, "Service components should be topologically sorted without errors")

	assert.Equal(t, len(c), 3, "Component topological sort should produce correct number of values")
	assert.Equal(t, "component1", c[0].Name, "Component topological sort should produce correct order")
	assert.Equal(t, "component2", c[1].Name, "Component topological sort should produce correct order")
	assert.Equal(t, "component3", c[2].Name, "Component topological sort should produce correct order")
}
