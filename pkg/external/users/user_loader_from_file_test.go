package users

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserLoaderFromFile(t *testing.T) {
	userLoader := NewUserLoaderFromFile("../../testdata/unittests/users.yaml", make(map[string]bool))

	// check user names
	users := userLoader.LoadUsersAll()
	assert.Equal(t, 7, len(users.Users), "Correct number of users should be loaded")

	names := []string{"Alice", "Bob", "Carol", "Dave", "Elena", "Sam", "Noname"}
	for _, name := range names {
		assert.Equal(t, name, users.Users[name].Name, "%s user should be loaded", name)
	}

	// check user labels
	userAlice := userLoader.LoadUserByName("Alice")
	assert.Equal(t, "Alice", userAlice.Name, "Alice should have correct name when loaded by ID")
	assert.Equal(t, 6, len(userAlice.Labels), "Alice should have correct label count")
	assert.Equal(t, "yes", userAlice.Labels["dev"], "Alice should have dev='yes' label")
	assert.Equal(t, "no", userAlice.Labels["prod"], "Alice should have prod='no' label")

	// check that summary is correct
	assert.Equal(t, "7 (from filesystem)", userLoader.Summary())
}
