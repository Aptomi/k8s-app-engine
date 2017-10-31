package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
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
	uMulti := NewUserLoaderMultipleSources([]UserLoader{u1, u2})
	assert.Equal(t, 25, len(uMulti.LoadUsersAll().Users), "Correct number of users should be loaded")
}
