package resolve

import (
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/builder"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ResSuccess = iota
	ResError   = iota
)

func resolvePolicy(t *testing.T, builder *builder.PolicyBuilder, expectedResult int, expectedLogMessage string) *PolicyResolution {
	t.Helper()
	resolver := NewPolicyResolver(builder.Policy(), builder.External())
	result, eventLog, err := resolver.ResolveAllDependencies()

	if !assert.Equal(t, expectedResult != ResError, err == nil, "Policy resolution status (success vs. error)") {
		// print log into stdout and exit
		hook := &event.HookStdout{}
		eventLog.Save(hook)
		t.FailNow()
		return nil
	}

	// check for error message
	verifier := event.NewUnitTestLogVerifier(expectedLogMessage, expectedResult == ResError)
	resolver.eventLog.Save(verifier)
	if !assert.True(t, verifier.MatchedErrorsCount() > 0, "Event log should have an error message containing words: "+expectedLogMessage) {
		hook := &event.HookStdout{}
		resolver.eventLog.Save(hook)
		t.FailNow()
	}

	return result
}

func getInstanceByDependencyKey(t *testing.T, dependencyID string, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := resolution.DependencyInstanceMap[dependencyID]
	if !assert.NotZero(t, len(key), "Dependency %s should be resolved", dependencyID) {
		t.Log(resolution.DependencyInstanceMap)
		t.FailNow()
	}
	instance, ok := resolution.ComponentInstanceMap[key]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key) {
		t.FailNow()
	}
	return instance
}

func getInstanceByParams(t *testing.T, cluster *lang.Cluster, contract *lang.Contract, context *lang.Context, allocationKeysResolved []string, service *lang.Service, component *lang.ServiceComponent, resolution *PolicyResolution) *ComponentInstance {
	t.Helper()
	key := NewComponentInstanceKey(cluster, contract, context, allocationKeysResolved, service, component)
	instance, ok := resolution.ComponentInstanceMap[key.GetKey()]
	if !assert.True(t, ok, "Component instance '%s' should be present in resolution data", key.GetKey()) {
		t.FailNow()
	}
	return instance
}
