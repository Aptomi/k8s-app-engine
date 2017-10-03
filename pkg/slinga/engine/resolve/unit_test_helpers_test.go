package resolve

import (
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func loadUnitTestsPolicy() *lang.Policy {
	return lang.LoadUnitTestsPolicy("../../testdata/unittests")
}

func loadPolicyAndResolve(t *testing.T) (*lang.Policy, *PolicyResolution) {
	t.Helper()
	policy := loadUnitTestsPolicy()
	return policy, resolvePolicy(t, policy, ResSuccess, "")
}

func resolvePolicy(t *testing.T, policy *lang.Policy, expectedResult int, expectedErrorMessage string) *PolicyResolution {
	t.Helper()
	externalData := external.NewData(
		users.NewUserLoaderFromDir("../../testdata/unittests"),
		secrets.NewSecretLoaderFromDir("../../testdata/unittests"),
	)
	resolver := NewPolicyResolver(policy, externalData)
	result, eventLog, err := resolver.ResolveAllDependencies()

	if !assert.Equal(t, expectedResult != ResError, err == nil, "Policy resolution status (success vs. error)") {
		// print log into stdout and exit
		hook := &event.HookStdout{}
		eventLog.Save(hook)
		t.FailNow()
		return nil
	}

	if expectedResult == ResError {
		// check for error message
		verifier := event.NewUnitTestLogVerifier(expectedErrorMessage)
		resolver.eventLog.Save(verifier)
		if !assert.True(t, verifier.MatchedErrorsCount() > 0, "Event log should have an error message containing words: "+expectedErrorMessage) {
			hook := &event.HookStdout{}
			resolver.eventLog.Save(hook)
			t.FailNow()
		}
		return nil
	}

	return result
}

func getInstanceByDependencyKey(t *testing.T, dependencyId string, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := resolution.DependencyInstanceMap[dependencyId]
	if !assert.NotZero(t, len(key), "Dependency %s should be resolved", dependencyId) {
		t.Log(resolution.DependencyInstanceMap)
		t.FailNow()
	}
	instance, ok := resolution.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, namespace string, clusterName string, contractName string, contextName string, allocationKeysResolved []string, componentName string, policy *lang.Policy, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	cluster := policy.Namespace[object.SystemNS].Clusters[clusterName]
	contract := policy.Namespace[namespace].Contracts[contractName]
	context := contract.FindContextByName(contextName)
	service := policy.Namespace[namespace].Services[context.Allocation.Service]
	key := NewComponentInstanceKey(cluster, contract, context, allocationKeysResolved, service, service.GetComponentsMap()[componentName])
	instance, ok := resolution.ComponentInstanceMap[key.GetKey()]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key.GetKey()) {
		t.FailNow()
	}
	return instance
}
