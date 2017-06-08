package slinga

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadUsers(t *testing.T) {
	users := LoadUsersFromDir("testdata/unittests")
	assert.Equal(t, 5, len(users.Users), "Two users should be loaded")
	assert.Equal(t, "Alice", users.Users["1"].Name, "Alice user should be loaded")
	assert.Equal(t, "Bob", users.Users["2"].Name, "Bob user should be loaded")
	assert.Equal(t, "Carol", users.Users["3"].Name, "Carol user should be loaded")
	assert.Equal(t, "Dave", users.Users["4"].Name, "Dave user should be loaded")
	assert.Equal(t, "Elena", users.Users["5"].Name, "Elena user should be loaded")

	userAlice := LoadUserByIDFromDir("testdata/unittests", "1")
	assert.Equal(t, "Alice", userAlice.Name, "Should load Alice user by ID")
	assert.Equal(t, "yes", userAlice.Labels["dev"], "Alice should have dev='yes' label")
	assert.Equal(t, "no", userAlice.Labels["prod"], "Alice should have prod='no' label")
}
