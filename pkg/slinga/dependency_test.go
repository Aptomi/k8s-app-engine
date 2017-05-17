package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadDependencies(t *testing.T) {
	dependencies := LoadDependenciesFromDir("testdata/fake")
	assert.Equal(t, 2, len(dependencies.Dependencies["kafka"]), "Service should have two dependencies")
}
