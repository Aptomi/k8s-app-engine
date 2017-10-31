package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
)

// UserLoader is an interface which allows aptomi to load user data from different sources (e.g. file, LDAP, AD, etc)
type UserLoader interface {
	// LoadUsersAll should load all users
	LoadUsersAll() *lang.GlobalUsers

	// LoadUserByName should load a single user by ID
	LoadUserByName(string) *lang.User

	// Summary returns summary for the loader as string
	Summary() string
}
