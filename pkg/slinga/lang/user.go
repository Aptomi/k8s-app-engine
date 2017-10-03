package lang

/*
	This file declares all the necessary structures for representing
	a User in Aptomi
*/

// User represents a user (ID, Name, set of labels)
type User struct {
	ID     string
	Name   string
	Labels map[string]string
}

// GlobalUsers contains the global list of users
type GlobalUsers struct {
	Users map[string]*User
}

// IsGlobalOps returns if user is a global ops guy
func (user *User) IsGlobalOps() bool {
	// TODO: this will have to be changed when we implement roles & ACLs
	return user.Labels["global_ops"] == "true"
}
