package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	policy := LoadUnitTestsPolicy("../testdata/unittests")

	// Check services
	assert.Equal(t, 2, len(policy.Services), "Two services should be loaded")
	assert.NotNil(t, policy.Services["kafka"], "Kafka service should be loaded")
	assert.NotNil(t, policy.Services["zookeeper"], "Zookeeper service should be loaded")

	// Check contracts
	assert.Equal(t, 2, len(policy.Contracts), "Two contracts should be loaded")
	assert.NotNil(t, policy.Contracts["kafka"], "Kafka contract should be loaded")
	assert.Equal(t, 3, len(policy.Contracts["kafka"].Contexts), "Kafka contract should have contexts")
	assert.NotNil(t, policy.Contracts["zookeeper"], "Zookeeper contract should be loaded")
	assert.Equal(t, 3, len(policy.Contracts["zookeeper"].Contexts), "Zookeeper contract should have contexts")

	// Check clusters
	assert.Equal(t, 2, len(policy.Clusters), "Two clusters should be loaded")

	// Check rules
	assert.Equal(t, 2, len(policy.Rules.Rules), "Correct number of rule action types should be loaded")

	// Check dependencies
	assert.Equal(t, 4, len(policy.Dependencies.DependenciesByContract["kafka"]), "Dependencies on kafka should be declared")
}
