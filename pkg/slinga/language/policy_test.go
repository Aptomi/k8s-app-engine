package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	state := LoadPolicyFromDir("../testdata/unittests")

	assert.Equal(t, 2, len(state.Services), "Two services should be loaded")

	assert.Equal(t, 2, len(state.Clusters), "Two clusters should be loaded")

	assert.Equal(t, 3, len(state.Contexts["kafka"]), "Three contexts should be loaded for kafka")
	assert.Equal(t, "kafka", state.Services["kafka"].Name, "Service name should be correct")
	assert.Equal(t, 3, len(state.Services["kafka"].Components), "Service should have components")
	assert.Equal(t, "prod", state.Contexts["kafka"][0].Name, "Context name should be correct")
	assert.NotNil(t, state.Contexts["kafka"][0].Allocation, "Context should have allocations")
	assert.NotNil(t, state.Contexts["kafka"][1].Allocation, "Context should have allocations")
	assert.NotNil(t, state.Contexts["kafka"][2].Allocation, "Context should have allocations")

	assert.Equal(t, 3, len(state.Contexts["zookeeper"]), "Three contexts should be loaded for zookeeper")
	assert.Equal(t, "zookeeper", state.Services["zookeeper"].Name, "Service name should be correct")
	assert.Equal(t, 2, len(state.Services["zookeeper"].Components), "Service should have components")
	assert.Equal(t, "prod", state.Contexts["zookeeper"][0].Name, "Context name should be correct")
	assert.NotNil(t, state.Contexts["zookeeper"][0].Allocation, "Context should have allocations")
	assert.NotNil(t, state.Contexts["zookeeper"][1].Allocation, "Context should have allocations")
	assert.NotNil(t, state.Contexts["zookeeper"][2].Allocation, "Context should have allocations")

	assert.Equal(t, "aptomi/code/unittests", state.Services["zookeeper"].Components[0].Code.Type, "ZooKeeper's first component should be unittests code")
	assert.Equal(t, "aptomi/code/unittests", state.Services["zookeeper"].Components[1].Code.Type, "ZooKeeper's second component should be unittests code")

	assert.Nil(t, state.Services["kafka"].Components[0].Code, "Kafka's first component should be service")
	assert.Equal(t, "zookeeper", state.Services["kafka"].Components[0].Service, "Kafka's first component should be service")
	assert.Equal(t, "aptomi/code/unittests", state.Services["kafka"].Components[1].Code.Type, "Kafka's second component should be unittests code")
	assert.Equal(t, "aptomi/code/unittests", state.Services["kafka"].Components[2].Code.Type, "Kafka's third component should be unittests code")
}
