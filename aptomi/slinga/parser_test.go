package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	state := loadGlobalState("testdata/")

	assert.Equal(t, 1, len(state.Services), "One service should be loaded");

	assert.Equal(t, 1, len(state.Contexts), "One context should be loaded");

	assert.Equal(t, "kafka", state.Services[0].Name, "Service name should be correct");

	assert.Equal(t, "test", state.Contexts[0].Name, "Context name should be correct");
}