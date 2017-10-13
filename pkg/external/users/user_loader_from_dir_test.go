package users

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadUsersFromDir(t *testing.T) {
	userLoader := NewUserLoaderFromDir("../../testdata/unittests")

	// check user names
	users := userLoader.LoadUsersAll()
	assert.Equal(t, 7, len(users.Users), "Correct number of users should be loaded")
	assert.Equal(t, "Alice", users.Users["1"].Name, "Alice user should be loaded")
	assert.Equal(t, "Bob", users.Users["2"].Name, "Bob user should be loaded")
	assert.Equal(t, "Carol", users.Users["3"].Name, "Carol user should be loaded")
	assert.Equal(t, "Dave", users.Users["4"].Name, "Dave user should be loaded")
	assert.Equal(t, "Elena", users.Users["5"].Name, "Elena user should be loaded")
	assert.Equal(t, "Sam", users.Users["6"].Name, "Sam user should be loaded")
	assert.Equal(t, "Noname", users.Users["7"].Name, "Sam user should be loaded")

	// check user labels
	userAlice := userLoader.LoadUserByID("1")
	assert.Equal(t, 6, len(userAlice.Labels), "Alice's labelset should have correct length")
	assert.Equal(t, "Alice", userAlice.Name, "Should load Alice user by ID")
	assert.Equal(t, "yes", userAlice.Labels["dev"], "Alice should have dev='yes' label")
	assert.Equal(t, "no", userAlice.Labels["prod"], "Alice should have prod='no' label")

	// check that summary is correct
	assert.Equal(t, "7 (from filesystem)", userLoader.Summary())
}
