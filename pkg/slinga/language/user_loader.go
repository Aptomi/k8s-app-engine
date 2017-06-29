package language

// UserLoader is an interface which allows aptomi to load users from different sources (e.g. file, LDAP, etc)
type UserLoader interface {
	// LoadUsersAll should load all users
	LoadUsersAll() GlobalUsers

	// LoadUserByID should load a single user by ID
	LoadUserByID(string) *User

	// Summary returns summary for the loader as string
	Summary() string
}

// NewAptomiUserLoader returns configured user loader for aptomi
func NewAptomiUserLoader() UserLoader {
	// return NewUserLoaderFromDir(GetAptomiPolicyDir())
	return NewUserLoaderFromLDAP()
}
