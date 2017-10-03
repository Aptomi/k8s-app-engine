package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	policy := LoadUnitTestsPolicy("../testdata/unittests")
	policyMain := policy.Namespace["main"]
	policySystem := policy.Namespace[object.SystemNS]

	// Check services
	assert.Equal(t, 2, len(policyMain.Services), "Two services should be loaded")
	assert.NotNil(t, policyMain.Services["kafka"], "Kafka service should be loaded")
	assert.NotNil(t, policyMain.Services["zookeeper"], "Zookeeper service should be loaded")

	// Check contracts
	assert.Equal(t, 2, len(policyMain.Contracts), "Two contracts should be loaded")
	assert.NotNil(t, policyMain.Contracts["kafka"], "Kafka contract should be loaded")
	assert.Equal(t, 3, len(policyMain.Contracts["kafka"].Contexts), "Kafka contract should have contexts")
	assert.NotNil(t, policyMain.Contracts["zookeeper"], "Zookeeper contract should be loaded")
	assert.Equal(t, 3, len(policyMain.Contracts["zookeeper"].Contexts), "Zookeeper contract should have contexts")

	// Check clusters
	assert.Equal(t, 2, len(policySystem.Clusters), "Two clusters should be loaded")

	// Check rules
	assert.Equal(t, 4, len(policyMain.Rules.Rules), "Correct number of rule action types should be loaded")

	// Check dependencies
	assert.Equal(t, 4, len(policyMain.Dependencies.DependenciesByContract["kafka"]), "Dependencies on kafka should be declared")
}
