package engine

import (
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)

func loadUnitTestsPolicy() *PolicyNamespace {
	policyLoader := NewSlingaObjectDatabaseDir("../testdata/unittests")
	policy := policyLoader.LoadPolicyObjects(-1, "")
	return policy
}

func loadPolicyAndResolve(t *testing.T) ServiceUsageState {
	return resolvePolicy(t, loadUnitTestsPolicy())
}

func resolvePolicy(t *testing.T, policy *PolicyNamespace) ServiceUsageState {
	userLoader := NewUserLoaderFromDir("../testdata/unittests")
	usageState := NewServiceUsageState(policy, userLoader)
	err := usageState.ResolveAllDependencies()
	if !assert.Nil(t, err, "PolicyNamespace usage should be resolved without errors") {
		t.FailNow()
	}
	return usageState
}

func emulateSaveAndLoadState(state ServiceUsageState) ServiceUsageState {
	// Emulate saving and loading again
	savedObjectAsString := yaml.SerializeObject(state)
	userLoader := NewUserLoaderFromDir("../testdata/unittests")
	loadedObject := ServiceUsageState{userLoader: userLoader}
	yaml.DeserializeObject(savedObjectAsString, &loadedObject)
	return loadedObject
}

func getInstance(t *testing.T, key string, usageData *ServiceUsageData) *ComponentInstance {
	instance, ok := usageData.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance in usage data: "+key) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, serviceName string, contextName string, allocationKeysResolved []string, componentName string, state ServiceUsageState) *ComponentInstance {
	key := NewComponentInstanceKey(serviceName, state.Policy.Contexts[contextName], allocationKeysResolved, state.Policy.Services[serviceName].GetComponentsMap()[componentName])
	return getInstance(t, key.GetKey(), state.ResolvedData)
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

func getTimes(t *testing.T, key string, u1 ServiceUsageState, u2 ServiceUsageState) componentTimes {
	return componentTimes{
		timePrevCreated: getInstance(t, key, u1.ResolvedData).CreatedOn,
		timePrevUpdated: getInstance(t, key, u1.ResolvedData).UpdatedOn,
		timeNextCreated: getInstance(t, key, u2.ResolvedData).CreatedOn,
		timeNextUpdated: getInstance(t, key, u2.ResolvedData).UpdatedOn,
	}
}

func getTimesNext(t *testing.T, key string, u2 ServiceUsageState) componentTimes {
	return componentTimes{
		timeNextCreated: getInstance(t, key, u2.ResolvedData).CreatedOn,
		timeNextUpdated: getInstance(t, key, u2.ResolvedData).UpdatedOn,
	}
}
