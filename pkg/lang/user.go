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
	Admin  bool
}

// GlobalUsers contains the global list of users
type GlobalUsers struct {
	Users map[string]*User
}
