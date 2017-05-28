package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadDependencies(t *testing.T) {
	dependencies := LoadDependenciesFromDir("testdata/unittests")
	assert.Equal(t, 4, len(dependencies.Dependencies["kafka"]), "Correct number of dependencies should be loaded")
	assert.Equal(t, 0, len(dependencies.Dependencies["kafka"][0].Labels), "First dependency should have 0 labels")
	assert.Equal(t, 1, len(dependencies.Dependencies["kafka"][1].Labels), "Second dependency should have 1 label")
	assert.Equal(t, 0, len(dependencies.Dependencies["kafka"][2].Labels), "Third dependency should have 0 labels")
	assert.Equal(t, 0, len(dependencies.Dependencies["kafka"][3].Labels), "Fourth dependency should have 0 labels")
}
