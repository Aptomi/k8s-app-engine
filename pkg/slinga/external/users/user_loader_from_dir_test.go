package users

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadUsersFromDir(t *testing.T) {
	userLoader := NewUserLoaderFromDir("../../testdata/unittests")

	users := userLoader.LoadUsersAll()

	// check user names
	assert.Equal(t, 7, len(users.Users), "Correct number of users should be loaded")
	assert.Equal(t, "Alice", users.Users["1"].Name, "Alice user should be loaded")
	assert.Equal(t, "Bob", users.Users["2"].Name, "Bob user should be loaded")
	assert.Equal(t, "Carol", users.Users["3"].Name, "Carol user should be loaded")
	assert.Equal(t, "Dave", users.Users["4"].Name, "Dave user should be loaded")
	assert.Equal(t, "Elena", users.Users["5"].Name, "Elena user should be loaded")
	assert.Equal(t, "Sam", users.Users["6"].Name, "Sam user should be loaded")
	assert.Equal(t, "Noname", users.Users["7"].Name, "Sam user should be loaded")

	userAlice := userLoader.LoadUserByID("1")
	userSam := userLoader.LoadUserByID("6")

	// check user labels
	assert.Equal(t, "Alice", userAlice.Name, "Should load Alice user by ID")
	assert.Equal(t, "yes", userAlice.Labels["dev"], "Alice should have dev='yes' label")
	assert.Equal(t, "no", userAlice.Labels["prod"], "Alice should have prod='no' label")
	assert.False(t, userAlice.IsGlobalOps(), "Alice should not be a global ops")
	assert.True(t, userSam.IsGlobalOps(), "Sam should be a global ops")

	// check user labels through a label set
	assert.Equal(t, 6, len(userAlice.GetLabelSet().Labels), "Alice's labelset should have correct length")
	assert.Equal(t, "yes", userAlice.GetLabelSet().Labels["dev"], "Alice should have dev='yes' label through a labelset")
	assert.Equal(t, "no", userAlice.GetLabelSet().Labels["prod"], "Alice should have prod='no' label through a labelset")

	// check that summary is correct
	assert.Equal(t, "7 (from filesystem)", userLoader.Summary())
}
