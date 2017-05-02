package slinga

import (
	"testing"
)

func TestEngine(t *testing.T) {
	state := loadGlobalStateFromDir("testdata/")
	users := loadUsersFromDir("testdata/")

	alice := users.Users["1"]
	state.resolve(alice, "kafka")
}
