package slinga

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoadUsers(t *testing.T) {
	users := loadUsersFromDir("testdata/")
	assert.Equal(t, 2, len(users.Users), "Two users should be loaded");
	assert.Equal(t, "Alice", users.Users["1"].Name, "Alice user should be loaded");
	assert.Equal(t, "Bob", users.Users["2"].Name, "Bob user should be loaded");

	userBob := loadUserByIDFromDir("testdata/", "2")
	assert.Equal(t, "Bob", userBob.Name, "Should load Bob user by ID");
}
