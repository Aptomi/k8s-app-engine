package users

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language"
)

// UserLoader is an interface which allows aptomi to load users from different sources (e.g. file, LDAP, etc)
type UserLoader interface {
	// LoadUsersAll should load all users
	LoadUsersAll() language.GlobalUsers

	// LoadUserByID should load a single user by ID
	LoadUserByID(string) *language.User

	// Summary returns summary for the loader as string
	Summary() string
}
