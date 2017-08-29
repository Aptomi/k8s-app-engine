package resolve

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestEnginePolicyResolutionAndResolvedData(t *testing.T) {
	policy, usageState := loadPolicyAndResolve(t)
	resolvedData := usageState.ResolvedData

	// Check that policy resolution finished correctly
	assert.Equal(t, 16, len(resolvedData.ComponentProcessingOrder), "Policy usage should have correct number of entries")

	// Resolution for test context
	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policy, usageState)
	assert.Equal(t, 1, len(kafkaTest.DependencyIds), "One dependency should be resolved with access to test")
	assert.True(t, policy.Dependencies.DependenciesByID["dep_id_1"].Resolved, "Only Alice should have access to test")

	// Resolution for prod context
	kafkaProd := getInstanceByParams(t, "kafka", "prod-low", []string{"team-platform_services", "true"}, "component2", policy, usageState)
	assert.Equal(t, 1, len(kafkaProd.DependencyIds), "One dependency should be resolved with access to prod")
	assert.Equal(t, "2", policy.Dependencies.DependenciesByID["dep_id_2"].UserID, "Only Bob should have access to prod (Carol is compromised)")
}

func TestEnginePolicyResolutionAndUnresolvedData(t *testing.T) {
	policy, _ := loadPolicyAndResolve(t)

	// Dave dependency on kafka should not be resolved
	daveOnKafkaDependency := policy.Dependencies.DependenciesByID["dep_id_4"]
	assert.False(t, daveOnKafkaDependency.Resolved, "Partial matching is broken. User has access to kafka, but not to zookeeper that kafka depends on. This should not be resolved successfully")
}

func TestEngineLabelProcessing(t *testing.T) {
	policy, usageState := loadPolicyAndResolve(t)

	// Check labels for Bob's dependency
	key := policy.Dependencies.DependenciesByID["dep_id_2"].ServiceKey
	serviceInstance := getInstanceInternal(t, key, usageState.ResolvedData)
	labels := serviceInstance.CalculatedLabels.Labels
	assert.Equal(t, "yes", labels["important"], "Label 'important=yes' should be carried from dependency all the way through the policy")
	assert.Equal(t, "true", labels["prod-low-ctx"], "Label 'prod-low-ctx=true' should be added on context match")
	assert.Equal(t, "", labels["some-label-to-be-removed"], "Label 'some-label-to-be-removed' should be removed on context match")
	assert.Equal(t, "true", labels["prod-low-alloc"], "Label 'prod-low-alloc=true' should be added on allocation match")
}

func TestEngineCodeAndDiscoveryParamsEval(t *testing.T) {
	policy, usageState := loadPolicyAndResolve(t)

	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", policy, usageState)

	// Check that code parameters evaluate correctly
	assert.Equal(t, "zookeeper-test-platform-services-component2", kafkaTest.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")

	// Check that discovery parameters evaluate correctly
	assert.Equal(t, "kafka-kafka-test-platform-services-component2-url", kafkaTest.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// Check that nested parameters evaluate correctly
	for i := 1; i <= 5; i++ {
		assert.Equal(t, "value"+strconv.Itoa(i), kafkaTest.CalculatedCodeParams.GetNestedMap("data" + strconv.Itoa(i)).GetNestedMap("param")["name"], "Nested code parameters should be calculated correctly")
	}
}
