package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	. "github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func BenchmarkEngine(b *testing.B) {
	t := &testing.T{}
	for i := 0; i < b.N; i++ {
		TestEngineComponentUpdateAndTimes(t)
	}
}

func TestEnginePolicyResolutionAndResolvedData(t *testing.T) {
	usageState := loadPolicyAndResolve(t)
	resolvedData := usageState.ResolvedData

	// Check that policy resolution finished correctly
	assert.Equal(t, 16, len(resolvedData.ComponentProcessingOrder), "Policy usage should have correct number of entries")

	// Resolution for test context
	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", usageState)
	assert.Equal(t, 1, len(kafkaTest.DependencyIds), "One dependency should be resolved with access to test")
	assert.True(t, usageState.Policy.Dependencies.DependenciesByID["dep_id_1"].Resolved, "Only Alice should have access to test")

	// Resolution for prod context
	kafkaProd := getInstanceByParams(t, "kafka", "prod-low", []string{"team-platform_services", "true"}, "component2", usageState)
	assert.Equal(t, 1, len(kafkaProd.DependencyIds), "One dependency should be resolved with access to prod")
	assert.Equal(t, "2", usageState.Policy.Dependencies.DependenciesByID["dep_id_2"].UserID, "Only Bob should have access to prod (Carol is compromised)")
}

func TestEnginePolicyResolutionAndUnresolvedData(t *testing.T) {
	usageState := loadPolicyAndResolve(t)

	// Dave dependency on kafka should not be resolved
	daveOnKafkaDependency := usageState.Policy.Dependencies.DependenciesByID["dep_id_4"]
	assert.False(t, daveOnKafkaDependency.Resolved, "Partial matching is broken. User has access to kafka, but not to zookeeper that kafka depends on. This should not be resolved successfully")
}

func TestEngineLabelProcessing(t *testing.T) {
	usageState := loadPolicyAndResolve(t)

	// Check labels for Bob's dependency
	key := usageState.Policy.Dependencies.DependenciesByID["dep_id_2"].ServiceKey
	serviceInstance := getInstance(t, key, usageState.ResolvedData)
	labels := serviceInstance.CalculatedLabels.Labels
	assert.Equal(t, "yes", labels["important"], "Label 'important=yes' should be carried from dependency all the way through the policy")
	assert.Equal(t, "true", labels["prod-low-ctx"], "Label 'prod-low-ctx=true' should be added on context match")
	assert.Equal(t, "", labels["some-label-to-be-removed"], "Label 'some-label-to-be-removed' should be removed on context match")
	assert.Equal(t, "true", labels["prod-low-alloc"], "Label 'prod-low-alloc=true' should be added on allocation match")
}

func TestEngineCodeAndDiscoveryParamsEval(t *testing.T) {
	usageState := loadPolicyAndResolve(t)

	kafkaTest := getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", usageState)

	// Check that code parameters evaluate correctly
	assert.Equal(t, "zookeeper-test-platform-services-component2", kafkaTest.CalculatedCodeParams["address"], "Code parameter should be calculated correctly")

	// Check that discovery parameters evaluate correctly
	assert.Equal(t, "kafka-kafka-test-platform-services-component2-url", kafkaTest.CalculatedDiscovery["url"], "Discovery parameter should be calculated correctly")

	// Check that nested parameters evaluate correctly
	for i := 1; i <= 5; i++ {
		assert.Equal(t, "value"+strconv.Itoa(i), kafkaTest.CalculatedCodeParams.GetNestedMap("data" + strconv.Itoa(i)).GetNestedMap("param")["name"], "Nested code parameters should be calculated correctly")
	}
}

func TestPolicyResolveEmptyDiff(t *testing.T) {
	usageStatePrev := loadPolicyAndResolve(t)
	usageStatePrev = emulateSaveAndLoadState(usageStatePrev)

	// Get usage state next
	usageStateNext := loadPolicyAndResolve(t)

	// Calculate and verify difference
	diff := usageStateNext.CalculateDifference(&usageStatePrev)
	verifyDiff(t, diff, false, 0, 0, 0, 0, 0)
}

func TestEngineNonEmptyDiffAndApplyNoop(t *testing.T) {
	usageStatePrev := loadPolicyAndResolve(t)
	usageStatePrev = emulateSaveAndLoadState(usageStatePrev)

	// Add another dependency and resolve usage state next
	policyNext := LoadUnitTestsPolicy()
	policyNext.Dependencies.AddDependency(
		&Dependency{
			Metadata: Metadata{
				Namespace: "main",
				Name:      "dep_id_5",
			},
			UserID:  "5",
			Service: "kafka",
		},
	)
	usageStateNext := resolvePolicy(t, policyNext)

	// Calculate difference
	diff := usageStateNext.CalculateDifference(&usageStatePrev)
	verifyDiff(t, diff, true, 8, 0, 0, 8, 0)

	// Apply diff in noop mode
	err := diff.Apply(true)
	assert.Nil(t, err, "Policy should be applied successfully in noop mode")
}

func TestEngineComponentUpdateAndTimes(t *testing.T) {
	var key string

	// Create initial empty state (do not resolve any dependencies)
	policyPrev := LoadUnitTestsPolicy()
	userLoader := NewUserLoaderFromDir("../testdata/unittests")
	uEmpty := NewServiceUsageState(policyPrev, userLoader)

	// Resolve all dependencies in policy
	uInitial := loadPolicyAndResolve(t)

	// Calculate difference against empty state to update times, emulate save/load
	uInitial.CalculateDifference(&uEmpty)
	uInitial = emulateSaveAndLoadState(uInitial)

	// Check creation/update times
	key = getInstanceByParams(t, "kafka", "test", []string{"platform_services"}, "component2", uInitial).Key.GetKey()
	p := getTimesNext(t, key, uInitial)
	assert.WithinDuration(t, time.Now(), p.timeNextCreated, time.Second, "Creation time should be initialized correctly for kafka")
	assert.Equal(t, p.timeNextUpdated, p.timeNextUpdated, "Update time should be equal to creation time")

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Add another dependency, resolve, calculate difference against prev state, emulate save/load
	policyNext := LoadUnitTestsPolicy()
	policyNext.Dependencies.AddDependency(
		&Dependency{
			Metadata: Metadata{
				Namespace: "main",
				Name:      "dep_id_5",
			},
			UserID:  "5",
			Service: "kafka",
		},
	)
	uNewDependency := resolvePolicy(t, policyNext)
	uNewDependency.CalculateDifference(&uInitial)

	// Check creation/update times
	pInit := getTimes(t, key, uInitial, uNewDependency)
	assert.Equal(t, pInit.timePrevCreated, pInit.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.Equal(t, pInit.timePrevUpdated, pInit.timeNextUpdated, "Update time should be carried over to remain the same")

	// Sleep a little bit to introduce time delay
	time.Sleep(25 * time.Millisecond)

	// Update user label, re-evaluate and see that component instance has changed
	userLoader.LoadUserByID("5").Labels["changinglabel"] = "newvalue"
	uUpdatedDependency := NewServiceUsageState(policyNext, userLoader)
	err := uUpdatedDependency.ResolveAllDependencies()
	assert.Nil(t, err, "All dependencies should be resolved successfully")

	diff := uUpdatedDependency.CalculateDifference(&uNewDependency)

	// Check that update has been performed
	verifyDiff(t, diff, true, 0, 0, 1, 0, 0)

	// Check creation/update times for component
	key = getInstanceByParams(t, "kafka", "prod-high", []string{"Elena"}, "component2", uNewDependency).Key.GetKey()
	pUpdate := getTimes(t, key, uNewDependency, uUpdatedDependency)
	assert.Equal(t, pUpdate.timePrevCreated, pUpdate.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.True(t, pUpdate.timeNextUpdated.After(pUpdate.timePrevUpdated), "Update time should be changed")

	// Check creation/update times for service
	key = getInstanceByParams(t, "kafka", "prod-high", []string{"Elena"}, componentRootName, uNewDependency).Key.GetKey()
	pUpdateSvc := getTimes(t, key, uNewDependency, uUpdatedDependency)
	assert.Equal(t, pUpdateSvc.timePrevCreated, pUpdateSvc.timeNextCreated, "Creation time should be carried over to remain the same")
	assert.True(t, pUpdateSvc.timeNextUpdated.After(pUpdateSvc.timePrevUpdated), "Update time should be changed for service")
}
