package slinga

import (
	log "github.com/Sirupsen/logrus"
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
	Users map[string]*User
}

// LoadUserByIDFromDir loads a given user from a given directory
func LoadUserByIDFromDir(dir string, id string) *User {
	return LoadUsersFromDir(dir).Users[id]
}

// LoadUsersFromDir loads all users from a given directory
func LoadUsersFromDir(dir string) GlobalUsers {
	fileName := dir + "/users.yaml"
	debug.WithFields(log.Fields{
		"file": fileName,
	}).Debug("Loading users")

	dat, e := ioutil.ReadFile(fileName)

	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to read file")
	}
	t := []*User{}
	e = yaml.Unmarshal([]byte(dat), &t)
	if e != nil {
		debug.WithFields(log.Fields{
			"file":  fileName,
			"error": e,
		}).Fatal("Unable to unmarshal users")
	}
	r := GlobalUsers{Users: make(map[string]*User)}
	for _, u := range t {
		r.Users[u.ID] = u
	}
	return r
}
