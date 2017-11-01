package lang

// User represents a user in Aptomi. It has a unique Name and a set of Labels,
// Users can be retrieved from multiple sources (e.g. file, LDAP, AD, etc)
type User struct {
	// Name is a unique name of a user
	Name string

	// Password hash is a hashed and salted user password
	PasswordHash string

	// Labels is a set of 'name'->'value' string labels, attached to the user
	Labels map[string]string

	// Admin is a special bool flag, which allows to mark certain users as domain admins. It's useful for Aptomi
	// bootstrap process, when someone needs to upload ACL rules into Aptomi (but his role is not defined in ACL,
	// because ACL list is empty when Aptomi is first installed)
	Admin bool
}

// GlobalUsers contains the map of users by their name
type GlobalUsers struct {
	// Users is a map[name] -> *User
	Users map[string]*User
}
