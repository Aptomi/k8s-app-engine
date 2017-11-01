package users

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"strconv"
	"strings"
)

// UserLoaderMultipleSources allows to combine different user sources into a single loader
type UserLoaderMultipleSources struct {
	loaders []UserLoader
	users   *lang.GlobalUsers
}

// NewUserLoaderMultipleSources returns new UserLoaderMultipleSources
func NewUserLoaderMultipleSources(loaders []UserLoader) *UserLoaderMultipleSources {
	result := &UserLoaderMultipleSources{
		loaders: loaders,
		users:   &lang.GlobalUsers{Users: make(map[string]*lang.User)},
	}
	for _, loader := range loaders {
		for name, user := range loader.LoadUsersAll().Users {
			result.users.Users[name] = user
		}
	}
	return result
}

// LoadUsersAll loads all users
func (loader *UserLoaderMultipleSources) LoadUsersAll() *lang.GlobalUsers {
	return loader.users
}

// LoadUserByName loads a single user by name
func (loader *UserLoaderMultipleSources) LoadUserByName(name string) *lang.User {
	name = strings.ToLower(name)
	return loader.users.Users[name]
}

// Authenticate authenticate a user by username/password by trying all available user data sources.
func (loader *UserLoaderMultipleSources) Authenticate(name, password string) (*lang.User, error) {
	for _, l := range loader.loaders {
		user := l.LoadUserByName(name)
		if user != nil {
			_, err := l.Authenticate(name, password)
			if err != nil {
				return nil, err
			}
			return user, err
		}
	}
	return nil, fmt.Errorf("user '%s' does not exist", name)
}

// Summary returns summary as string
func (loader *UserLoaderMultipleSources) Summary() string {
	return strconv.Itoa(len(loader.users.Users)) + " (multiple sources)"
}
