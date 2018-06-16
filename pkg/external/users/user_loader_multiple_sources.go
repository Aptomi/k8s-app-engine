package users

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Aptomi/aptomi/pkg/lang"
)

// UserLoaderMultipleSources allows to combine different user sources into a single loader
type UserLoaderMultipleSources struct {
	loaders []UserLoader
}

// NewUserLoaderMultipleSources returns new UserLoaderMultipleSources
func NewUserLoaderMultipleSources(loaders []UserLoader) *UserLoaderMultipleSources {
	return &UserLoaderMultipleSources{loaders: loaders}
}

// LoadUsersAll loads all users
func (loader *UserLoaderMultipleSources) LoadUsersAll() *lang.GlobalUsers {
	result := &lang.GlobalUsers{Users: make(map[string]*lang.User)}
	for _, loader := range loader.loaders {
		for name, user := range loader.LoadUsersAll().Users {
			result.Users[strings.ToLower(name)] = user
		}
	}
	return result
}

// LoadUserByName loads a single user by name
func (loader *UserLoaderMultipleSources) LoadUserByName(name string) *lang.User {
	for _, l := range loader.loaders {
		user := l.LoadUserByName(name)
		if user != nil {
			return user
		}
	}
	return nil
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
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (multiple sources)"
}
