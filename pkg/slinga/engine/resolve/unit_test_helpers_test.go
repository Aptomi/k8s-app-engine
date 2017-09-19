package resolve

import (
	"github.com/Aptomi/aptomi/pkg/slinga/eventlog"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	. "github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func loadUnitTestsPolicy() *Policy {
	return LoadUnitTestsPolicy("../../testdata/unittests")
}

func loadPolicyAndResolve(t *testing.T) (*Policy, *PolicyResolution) {
	t.Helper()
	policy := loadUnitTestsPolicy()
	return policy, resolvePolicy(t, policy, ResSuccess, "")
}

func resolvePolicy(t *testing.T, policy *Policy, expectedResult int, expectedErrorMessage string) *PolicyResolution {
	t.Helper()
	externalData := external.NewData(
		users.NewUserLoaderFromDir("../../testdata/unittests"),
		secrets.NewSecretLoaderFromDir("../../testdata/unittests"),
	)
	resolver := NewPolicyResolver(policy, externalData)
	result, eventLog, err := resolver.ResolveAllDependencies()

	if !assert.Equal(t, expectedResult != ResError, err == nil, "Policy resolution status (success vs. error)") {
		// print log into stdout and exit
		hook := &eventlog.HookStdout{}
		eventLog.Save(hook)
		t.FailNow()
		return nil
	}

	if expectedResult == ResError {
		// check for error message
		verifier := eventlog.NewUnitTestLogVerifier(expectedErrorMessage)
		resolver.eventLog.Save(verifier)
		if !assert.True(t, verifier.MatchedErrorsCount() > 0, "Event log should have an error message containing words: "+expectedErrorMessage) {
			hook := &eventlog.HookStdout{}
			resolver.eventLog.Save(hook)
			t.FailNow()
		}
		return nil
	}

	return result
}

func getInstanceByDependencyId(t *testing.T, dependencyId string, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := resolution.DependencyInstanceMap[dependencyId]
	if !assert.NotZero(t, len(key), "Dependency %s should be resolved", dependencyId) {
		t.FailNow()
	}
	instance, ok := resolution.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, clusterName string, contractName string, contextName string, allocationKeysResolved []string, componentName string, policy *Policy, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	cluster := policy.Clusters[clusterName]
	contract := policy.Contracts[contractName]
	context := contract.FindContextByName(contextName)
	service := policy.Services[context.Allocation.Service]
	key := NewComponentInstanceKey(cluster, contract, context, allocationKeysResolved, service, service.GetComponentsMap()[componentName])
	instance, ok := resolution.ComponentInstanceMap[key.GetKey()]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key.GetKey()) {
		t.FailNow()
	}
	return instance
}
