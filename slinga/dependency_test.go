package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoadDependencies(t *testing.T) {
	dependencies := loadDependenciesFromDir("testdata/")
	assert.Equal(t, 2, len(dependencies.Dependencies["kafka"]), "Service should have two dependencies");
}
