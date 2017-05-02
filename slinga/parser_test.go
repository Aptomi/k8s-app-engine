package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	state := loadGlobalStateFromDir("testdata/")

	assert.Equal(t, 1, len(state.Services), "One service should be loaded");
	assert.Equal(t, 2, len(state.Contexts["kafka"]), "Two contexts should be loaded");
	assert.Equal(t, "kafka", state.Services["kafka"].Name, "Service name should be correct");
	assert.Equal(t, 3, len(state.Services["kafka"].Components), "Service should have components");
	assert.Equal(t, "prod", state.Contexts["kafka"][0].Name, "Context name should be correct");
	assert.Equal(t, 2, len(state.Contexts["kafka"][0].Allocations), "Context should have allocations");
}
