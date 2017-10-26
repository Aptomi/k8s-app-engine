package lang

// User represents a user in Aptomi. It has an ID, Name and a set of Labels,
// Users can be retrieved from multiple sources (e.g. file for Aptomi bootstrap, LDAP, AD, etc)
type User struct {
	// ID is a unique string identifier for a user
	ID string

	// Name is a name of a user
	Name string

	// Labels is a set of 'name'->'value' string labels, attached to the user
	Labels map[string]string

	// Admin is a special bool flag, which allows to mark certain users as domain admins. It's useful in Aptomi
	// bootstrap process, when someone needs to upload ACL rules into Aptomi (but his role is not defined in ACL,
	// because ACL list is empty when Aptomi is first installed)
	Admin bool
}

// GlobalUsers contains the map of users by ID
type GlobalUsers struct {
	// Users is a map[ID] -> *User
	Users map[string]*User
}
