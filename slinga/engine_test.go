package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"log"
)

func TestEngine(t *testing.T) {
	state := loadGlobalStateFromDir("testdata/")
	users := loadUsersFromDir("testdata/")

	alice := users.Users["1"]
	_, err := state.resolve(alice, "kafka")
	if err != nil {
		log.Fatal(err)
	}

	// TODO:
}

func TestServiceComponentsTopologicalOrder(t *testing.T) {
	state := loadGlobalStateFromDir("testdata/")
	service := state.Services["kafka"]

	err := service.sortComponentsTopologically()
	assert.Equal(t, nil, err, "Service components should be topologically sorted without errors")

	assert.Equal(t, "component3", service.ComponentsOrdered[0].Name, "Component tologogical sort should produce correct order")
	assert.Equal(t, "component2", service.ComponentsOrdered[1].Name, "Component tologogical sort should produce correct order")
	assert.Equal(t, "component1", service.ComponentsOrdered[2].Name, "Component tologogical sort should produce correct order")
}
