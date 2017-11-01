package users

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestUserLoaderFromFile(t *testing.T) {
	userLoader := NewUserLoaderFromFile("../../testdata/unittests/users.yaml", make(map[string]bool))

	// check user names
	users := userLoader.LoadUsersAll()
	assert.Equal(t, 3, len(users.Users), "Correct number of users should be loaded")

	names := []string{"alice", "bob", "carol"}
	for _, name := range names {
		assert.Equal(t, name, users.Users[name].Name, "%s user should be loaded", name)
	}

	// load user by name (not case sensitive). check user labels
	userAlice := userLoader.LoadUserByName("Alice")
	assert.Equal(t, 6, len(userAlice.Labels), "Alice should have correct label count")
	assert.Equal(t, "yes", userAlice.Labels["dev"], "Alice should have dev='yes' label")
	assert.Equal(t, "no", userAlice.Labels["prod"], "Alice should have prod='no' label")

	// check that summary is correct
	assert.Equal(t, "3 (from filesystem)", userLoader.Summary())

	// test that authentication works
	for _, name := range names {
		// successful authentication (password == lowercase name for test users, e.g. 'alice')
		{
			user, err := userLoader.Authenticate(name, strings.ToLower(name))
			assert.NoError(t, err, "Authentication should be successful")
			assert.NotEmpty(t, user, "User should be returned as a result of authentication")
		}
		// successful authentication (user names are not case sensitive)
		{
			user, err := userLoader.Authenticate(strings.ToUpper(name), strings.ToLower(name))
			assert.NoError(t, err, "Authentication should be successful (user name not case sensitive)")
			assert.NotEmpty(t, user, "User should be returned as a result of authentication")
		}
		// failed authentication
		{
			user, err := userLoader.Authenticate(name, name+"pass")
			assert.Error(t, err, "Authentication should not be successful")
			assert.Empty(t, user, "User should not be returned as a result of failed authentication")
		}
	}
}
