package slinga

import (
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

/*
	This file declares all the necessary structures for Users to be retrieved
	For now it loads users with labels from a YAML file
	Later this will be replaced with LDAP integration
*/

// User represents a user (ID, Name, set of labels)
type User struct {
	ID     string
	Name   string
	Labels map[string]string
}

// GlobalUsers contains the global list of users
type GlobalUsers struct {
	Users map[string]User
}

// LoadUserByIDFromDir loads a given user from a given directory
func LoadUserByIDFromDir(dir string, id string) User {
	return LoadUsersFromDir(dir).Users[id]
}

// LoadUsersFromDir loads all users from a given directory
func LoadUsersFromDir(dir string) GlobalUsers {
	dat, e := ioutil.ReadFile(dir + "/users.yaml")
	if e != nil {
		glog.Fatalf("Unable to read file: %v", e)
	}
	t := []User{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		glog.Fatalf("Unable to unmarshal user: %v", e)
	}
	r := GlobalUsers{Users: make(map[string]User)}
	for _, u := range t {
		r.Users[u.ID] = u
	}
	return r
}
