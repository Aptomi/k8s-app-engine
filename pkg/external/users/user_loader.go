package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
)

// UserLoader is an interface which allows aptomi to load user data from different sources (e.g. file, LDAP, AD, etc)
type UserLoader interface {
	// LoadUsersAll should load all users
	LoadUsersAll() *lang.GlobalUsers

	// LoadUserByName should load a single user by name. Name is not case sensitive.
	LoadUserByName(name string) *lang.User

	// Authenticate should authenticate a user by username/password. Name is not case sensitive.
	// If user exists and username/password is correct, it should be returned.
	// If a user doesn't exist or username/password is not correct, then nil should be returned together with an error.
	Authenticate(name, password string) (*lang.User, error)

	// Summary returns summary for the loader as string
	Summary() string
}
