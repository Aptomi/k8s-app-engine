package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadDependencies(t *testing.T) {
	dependencies := LoadUnitTestsPolicy("../testdata/unittests").Dependencies

	// look at kafka dependencies
	kafkaDeps := dependencies.DependenciesByService["kafka"]
	assert.Equal(t, 4, len(kafkaDeps), "Correct number of dependencies should be loaded")
	assert.Equal(t, "dep_id_1", kafkaDeps[0].GetID(), "Dependency ID should be correct")
	assert.Equal(t, 0, len(kafkaDeps[0].Labels), "First dependency should have 0 labels")
	assert.Equal(t, 2, len(kafkaDeps[1].Labels), "Second dependency should have 1 label")
	assert.Equal(t, 0, len(kafkaDeps[2].Labels), "Third dependency should have 0 labels")
	assert.Equal(t, 0, len(kafkaDeps[3].Labels), "Fourth dependency should have 0 labels")

	// look at dependency labels for 'dep_id_2'
	dep2 := dependencies.DependenciesByID["dep_id_2"]
	assert.Equal(t, 2, len(dep2.Labels), "dep_id_2's labelset should have correct length")
	assert.Equal(t, "yes", dep2.Labels["important"], "dep_id_2's should have important='yes' label through a labelset")
	assert.Equal(t, "yes", dep2.Labels["some-label-to-be-removed"], "dep_id_2's should have some-label-to-be-removed='yes' label through a labelset")
}

func TestAddDependency(t *testing.T) {
	dependencies := LoadUnitTestsPolicy("../testdata/unittests").Dependencies

	// create new dependency
	depAdd := &Dependency{
		Metadata: Metadata{
			Namespace: "main",
			Name:      "dep_id_new",
		},
		UserID:  "612",
		Service: "newservice",
	}

	// add it to the list of global dependencies
	dependencies.AddDependency(depAdd)

	// check that it was successfully added
	assert.Equal(t, 1, len(dependencies.DependenciesByService["newservice"]), "Dependency on 'newservice' should be added")
	assert.Equal(t, "dep_id_new", dependencies.DependenciesByService["newservice"][0].Name, "Dependency on 'newservice' should be added")
	assert.Equal(t, "dep_id_new", dependencies.DependenciesByID["dep_id_new"].Name, "Dependency on 'newservice' should be added")
}

func TestAddInvalidDependency(t *testing.T) {
	dependencies := NewGlobalDependencies()

	depAdd := &Dependency{
		Metadata: Metadata{
			Namespace: "main",
			Name:      "",
		},
		UserID:  "612",
		Service: "newservice",
	}

	assert.Panics(t, assert.PanicTestFunc(func() { dependencies.AddDependency(depAdd) }), "Adding invalid dependency should result in panic")
}
