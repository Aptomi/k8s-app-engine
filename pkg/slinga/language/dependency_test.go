package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadDependencies(t *testing.T) {
	dependencies := LoadDependenciesFromDir("../testdata/unittests")
	assert.Equal(t, 4, len(dependencies.DependenciesByService["kafka"]), "Correct number of dependencies should be loaded")
	assert.Equal(t, "dep_id_1", dependencies.DependenciesByService["kafka"][0].ID, "Dependency ID should be correct")
	assert.Equal(t, 0, len(dependencies.DependenciesByService["kafka"][0].Labels), "First dependency should have 0 labels")
	assert.Equal(t, 1, len(dependencies.DependenciesByService["kafka"][1].Labels), "Second dependency should have 1 label")
	assert.Equal(t, 0, len(dependencies.DependenciesByService["kafka"][2].Labels), "Third dependency should have 0 labels")
	assert.Equal(t, 0, len(dependencies.DependenciesByService["kafka"][3].Labels), "Fourth dependency should have 0 labels")
}
