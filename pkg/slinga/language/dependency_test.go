package language

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
	dependencies.AddDependency(depAdd)

	// check that it was successfully added
	assert.Equal(t, 1, len(dependencies.DependenciesByContract["newcontract"]), "Dependency on 'newcontract' should be added")
	assert.Equal(t, "dep_id_new", dependencies.DependenciesByContract["newcontract"][0].Name, "Dependency on 'newcontract' should be added")
	assert.Equal(t, "dep_id_new", dependencies.DependenciesByID["dep_id_new"].Name, "Dependency on 'newcontract' should be added")
}

func TestAddInvalidDependency(t *testing.T) {
	dependencies := NewGlobalDependencies()

	depAdd := &Dependency{
		Metadata: Metadata{
			Namespace: "main",
			Name:      "",
		},
		UserID:   "612",
		Contract: "newcontract",
	}

	assert.Panics(t, assert.PanicTestFunc(func() { dependencies.AddDependency(depAdd) }), "Adding invalid dependency should result in panic")
}
