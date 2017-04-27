package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	state := loadGlobalStateFromDir("testdata/")

	assert.Equal(t, 1, len(state.Services), "One service should be loaded");

	assert.Equal(t, 1, len(state.Contexts), "One context should be loaded");

	assert.Equal(t, "kafka", state.Services[0].Name, "Service name should be correct");

	assert.Equal(t, 2, len(state.Services[0].Components), "Service should have components");

	assert.Equal(t, "test", state.Contexts[0].Name, "Context name should be correct");

	assert.Equal(t, 2, len(state.Contexts[0].Allocations), "Context should have allocations");
}
