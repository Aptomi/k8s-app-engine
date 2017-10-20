package lang

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddDependency(t *testing.T) {
	dependencies := NewGlobalDependencies()

	// create new dependency
	depAdd := &Dependency{
		Metadata: Metadata{
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:   "612",
		Contract: "newcontract",
	}

	// add it to the list of global dependencies
	dependencies.addDependency(depAdd)

	// check that it was successfully added
	assert.Equal(t, 1, len(dependencies.DependenciesByContract["newcontract"]), "Dependency on 'newcontract' should be added")
	assert.Equal(t, "dep_id_new", dependencies.DependenciesByContract["newcontract"][0].Name, "Dependency on 'newcontract' should be added")
}
