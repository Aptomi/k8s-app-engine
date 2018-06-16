package users

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/stretchr/testify/assert"
)

func makeUserLoader(offset, users int) UserLoader {
	loader := NewUserLoaderMock()
	for i := 0; i < users; i++ {
		loader.AddUser(&lang.User{
			Name: strconv.Itoa(i + offset),
		})
	}
	return loader
}

func TestUserLoaderFromMultipleSources(t *testing.T) {
	u1 := makeUserLoader(0, 10)
	u2 := makeUserLoader(len(u1.LoadUsersAll().Users), 15)
	unitTests := NewUserLoaderFromFile("../../testdata/unittests/users.yaml", make(map[string]bool))
	uMulti := NewUserLoaderMultipleSources([]UserLoader{u1, u2, unitTests})

	// Test that LoadUsersAll() works correctly
	assert.Equal(t, 28, len(uMulti.LoadUsersAll().Users), "Correct number of users should be loaded")

	// Test that LoadUserByName() works correctly
	alice := uMulti.LoadUserByName("aLICe")
	assert.NotNil(t, alice, "Alice should be loaded by name")
	assert.Equal(t, "Alice", alice.Name, "Alice should be loaded by name")

	nonExisting := uMulti.LoadUserByName("not-existing")
	assert.Nil(t, nonExisting, "Non-existing user should be nil when loaded ny name")

	// Test that Authenticate() works correctly
	names := []string{"Alice", "Bob", "Carol"}
	for _, name := range names {
		// successful authentication (password == lowercase name for test users, e.g. 'alice')
		{
			user, err := uMulti.Authenticate(name, strings.ToLower(name))
			assert.NoError(t, err, "Authentication should be successful")
			assert.NotEmpty(t, user, "User should be returned as a result of authentication")
		}
		// failed authentication (existing user, incorrect password)
		{
			user, err := uMulti.Authenticate(name, name+"pass")
			assert.Error(t, err, "Authentication should not be successful")
			assert.Empty(t, user, "User should not be returned as a result of failed authentication")
		}
	}
	// failed authentication with non-existing user
	{
		user, err := uMulti.Authenticate("non-existing", "non-existing")
		assert.Error(t, err, "Authentication should not be successful")
		assert.Empty(t, user, "User should not be returned as a result of failed authentication")
	}
}
