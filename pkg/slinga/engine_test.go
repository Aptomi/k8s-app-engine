package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"gopkg.in/yaml.v2"
)

func TestPolicyResolve(t *testing.T) {
	policy := LoadPolicyFromDir("testdata/unittests")
	users := LoadUsersFromDir("testdata/unittests")
	dependencies := LoadDependenciesFromDir("testdata/unittests")

	usageState := NewServiceUsageState(&policy, &dependencies, &users)
	err := usageState.ResolveAllDependencies()
	resolvedUsage := usageState.getResolvedUsage()

	// Check that policy resolution finished correctly
	assert.Nil(t, err, "Policy usage should be resolved without errors")
	assert.NotEqual(t, 0, len(resolvedUsage.ComponentProcessingOrder), "Policy usage should have entries")

	kafkaTest := resolvedUsage.ComponentInstanceMap["kafka#test#test-platform_services#component2"]
	kafkaProd := resolvedUsage.ComponentInstanceMap["kafka#prod#prod-platform_services#component2"]
	assert.Equal(t, 1, len(kafkaTest.UserIds), "Only one user should have access to test")
	assert.Equal(t, "1", kafkaTest.UserIds[0], "Only Alice should have access to test")

	assert.Equal(t, 1, len(kafkaProd.UserIds), "Only one user should have access to prod")
	assert.Equal(t, "2", kafkaProd.UserIds[0], "Only Bob should have access to prod (Carol is compromised)")

	// Check that code parameters evaluate correctly
	assert.Equal(t, "zookeeper-test-test-platform-services-component2", kafkaTest.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")

	// Check that discovery parameters evaluate correctly
	assert.Equal(t, "kafka-kafka-test-test-platform-services-component2-url", kafkaTest.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")
}

func TestPolicyResolveEmptyDiff(t *testing.T) {
	policy := LoadPolicyFromDir("testdata/unittests")
	users := LoadUsersFromDir("testdata/unittests")
	dependencies := LoadDependenciesFromDir("testdata/unittests")

	// Get usage state prev
	usageStatePrev := NewServiceUsageState(&policy, &dependencies, &users)
	usageStatePrev.ResolveAllDependencies()

	// Emulate saving and loading again
	usageStatePrevSavedLoaded := ServiceUsageState{}
	yaml.Unmarshal([]byte(serializeObject(usageStatePrev)), &usageStatePrevSavedLoaded)

	// Get usage state next
	usageStateNext := NewServiceUsageState(&policy, &dependencies, &users)
	usageStateNext.ResolveAllDependencies()

	// Calculate difference
	diff := usageStateNext.CalculateDifference(&usageStatePrevSavedLoaded)

	assert.Equal(t, 0, len(diff.ComponentInstantiate), "Empty diff should not have any component instantiations")
	assert.Equal(t, 0, len(diff.ComponentDestruct), "Empty diff should not have any component destructions")
	assert.Equal(t, 0, len(diff.ComponentUpdate), "Empty diff should not have any component updates")
	assert.Equal(t, 0, len(diff.ComponentAttachUser), "Empty diff should not have any users attached to components")
	assert.Equal(t, 0, len(diff.ComponentDetachUser), "Empty diff should not have any users removed from components")
}

func TestPolicyResolveNonEmptyDiff(t *testing.T) {
	policy := LoadPolicyFromDir("testdata/unittests")
	users := LoadUsersFromDir("testdata/unittests")
	dependenciesPrev := LoadDependenciesFromDir("testdata/unittests")

	// Get usage state prev
	usageStatePrev := NewServiceUsageState(&policy, &dependenciesPrev, &users)
	usageStatePrev.ResolveAllDependencies()

	// Emulate saving and loading again
	usageStatePrevSavedLoaded := ServiceUsageState{}
	yaml.Unmarshal([]byte(serializeObject(usageStatePrev)), &usageStatePrevSavedLoaded)

	// Add another dependency and resolve usage state next
	dependenciesNext := dependenciesPrev.appendDependency(
		&Dependency{
			UserID:  "5",
			Service: "kafka",
		},
	)
	usageStateNext := NewServiceUsageState(&policy, &dependenciesNext, &users)
	usageStateNext.ResolveAllDependencies()

	// Calculate difference
	diff := usageStateNext.CalculateDifference(&usageStatePrevSavedLoaded)

	assert.Equal(t, 7, len(diff.ComponentInstantiate), "Diff should have component instantiations")
	assert.Equal(t, 0, len(diff.ComponentDestruct), "Diff should not have any component destructions")
	assert.Equal(t, 0, len(diff.ComponentUpdate), "Diff should not have any component updates")
	assert.Equal(t, 7, len(diff.ComponentAttachUser), "Diff should not have any users attached to components")
	assert.Equal(t, 0, len(diff.ComponentDetachUser), "Diff should not have any users removed from components")
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
