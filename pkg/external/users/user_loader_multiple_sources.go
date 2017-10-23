package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"strconv"
)

// UserLoaderMultipleSources allows to combine different user sources into a single loader
type UserLoaderMultipleSources struct {
	users *lang.GlobalUsers
}

// NewUserLoaderMultipleSources returns new UserLoaderMultipleSources
func NewUserLoaderMultipleSources(loaders []UserLoader) *UserLoaderMultipleSources {
	result := &UserLoaderMultipleSources{users: &lang.GlobalUsers{Users: make(map[string]*lang.User)}}
	for _, loader := range loaders {
		for userID, user := range loader.LoadUsersAll().Users {
			result.users.Users[userID] = user
		}
	}
	return result
}

// LoadUsersAll loads all users
func (loader *UserLoaderMultipleSources) LoadUsersAll() *lang.GlobalUsers {
	return loader.users
}

// LoadUserByID loads a single user by ID
func (loader *UserLoaderMultipleSources) LoadUserByID(id string) *lang.User {
	return loader.users.Users[id]
}

// Summary returns summary as string
func (loader *UserLoaderMultipleSources) Summary() string {
	return strconv.Itoa(len(loader.users.Users)) + " (multiple sources)"
}
