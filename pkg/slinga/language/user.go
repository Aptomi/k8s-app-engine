package language

/*
	This file declares all the necessary structures for Users to be retrieved
	For now it loads users with labels from a YAML file
	Later this will be replaced with LDAP integration
*/

// User represents a user (ID, Name, set of labels)
type User struct {
	ID      string
	Name    string
	Labels  map[string]string
	Secrets map[string]string
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

// GetLabelSet returns a set of user labels
func (user *User) GetLabelSet() LabelSet {
	return NewLabelSet(user.Labels)
}

// GetSecretSet returns a set of user secrets
func (user *User) GetSecretSet() LabelSet {
	return NewLabelSetSecret(user.Secrets)
}
