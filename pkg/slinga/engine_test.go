package slinga

import (
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEngine(t *testing.T) {
	policy := LoadPolicyFromDir("testdata/fake")
	users := LoadUsersFromDir("testdata/fake")
	dependencies := LoadDependenciesFromDir("testdata/fake")

	usageState := NewServiceUsageState(&policy, &dependencies)
	err := usageState.ResolveUsage(&users)

	if err != nil {
		glog.Fatal(err)
	}

	assert.Equal(t, 14, len(usageState.ResolvedLinks), "Policy resolution should result in correct amount of usage entries")
}

func TestServiceComponentsTopologicalOrder(t *testing.T) {
	state := LoadPolicyFromDir("testdata/fake")
	service := state.Services["kafka"]

	err := service.sortComponentsTopologically()
	assert.Equal(t, nil, err, "Service components should be topologically sorted without errors")

	assert.Equal(t, "component3", service.ComponentsOrdered[0].Name, "Component tologogical sort should produce correct order")
	assert.Equal(t, "component2", service.ComponentsOrdered[1].Name, "Component tologogical sort should produce correct order")
	assert.Equal(t, "component1", service.ComponentsOrdered[2].Name, "Component tologogical sort should produce correct order")
}
