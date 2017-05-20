package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadDependencies(t *testing.T) {
	dependencies := LoadDependenciesFromDir("testdata/unittests")
	assert.Equal(t, 2, len(dependencies.Dependencies["kafka"]), "Service should have two dependencies")
	assert.Equal(t, 0, len(dependencies.Dependencies["kafka"][0].Labels), "First dependency should have 0 labels")
	assert.Equal(t, 1, len(dependencies.Dependencies["kafka"][1].Labels), "Second dependency should have 1 label")
}
