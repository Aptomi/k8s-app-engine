package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	state := LoadPolicyFromDir("testdata/fake")

	assert.Equal(t, 2, len(state.Services), "Two services should be loaded");

	assert.Equal(t, 2, len(state.Contexts["kafka"]), "Two contexts should be loaded for kafka");
	assert.Equal(t, "kafka", state.Services["kafka"].Name, "Service name should be correct");
	assert.Equal(t, 3, len(state.Services["kafka"].Components), "Service should have components");
	assert.Equal(t, "prod", state.Contexts["kafka"][0].Name, "Context name should be correct");
	assert.Equal(t, 2, len(state.Contexts["kafka"][0].Allocations), "Context should have allocations");

	assert.Equal(t, 2, len(state.Contexts["zookeeper"]), "Two contexts should be loaded for zookeeper");
	assert.Equal(t, "zookeeper", state.Services["zookeeper"].Name, "Service name should be correct");
	assert.Equal(t, 2, len(state.Services["zookeeper"].Components), "Service should have components");
	assert.Equal(t, "prod", state.Contexts["zookeeper"][0].Name, "Context name should be correct");
	assert.Equal(t, 2, len(state.Contexts["zookeeper"][0].Allocations), "Context should have allocations");

}
