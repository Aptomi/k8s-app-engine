package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	policy := LoadUnitTestsPolicy("../testdata/unittests")

	// Check services
	assert.Equal(t, 4, len(policy.Services), "Two services should be loaded")
	assert.Equal(t, "kafka", policy.Services["kafka"].Name, "Service name should be correct")
	assert.Equal(t, 4, len(policy.Services["kafka"].Components), "Service should have components")

	// Check clusters
	assert.Equal(t, 2, len(policy.Clusters), "Two clusters should be loaded")

	// Check contexts
	assert.Equal(t, 8, len(policy.Contexts), "Five contexts should be loaded")
	assert.Equal(t, "test", policy.Contexts["test"].Name, "Context name should be correct")
	assert.NotNil(t, policy.Contexts["prod-high"].Allocation, "Context should have allocations")
	assert.NotNil(t, policy.Contexts["prod-low"].Allocation, "Context should have allocations")
	assert.NotNil(t, policy.Contexts["test"].Allocation, "Context should have allocations")

	assert.Equal(t, "aptomi/code/unittests", policy.Services["zookeeper"].Components[0].Code.Type, "ZooKeeper's first component should be unittests code")
	assert.Equal(t, "aptomi/code/unittests", policy.Services["zookeeper"].Components[1].Code.Type, "ZooKeeper's second component should be unittests code")

	assert.Nil(t, policy.Services["kafka"].Components[0].Code, "Kafka's first component should be service")
	assert.Equal(t, "zookeeper", policy.Services["kafka"].Components[0].Service, "Kafka's first component should be service")
	assert.Equal(t, "aptomi/code/unittests", policy.Services["kafka"].Components[1].Code.Type, "Kafka's second component should be unittests code")
	assert.Equal(t, "aptomi/code/unittests", policy.Services["kafka"].Components[2].Code.Type, "Kafka's third component should be unittests code")
}
