package slinga

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
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
	ID      string
	Name    string
	Labels  map[string]string
	Secrets map[string]string
}

// GlobalUsers contains the global list of users
type GlobalUsers struct {
	Users map[string]*User
}

func (users *GlobalUsers) count() int {
	return countElements(users.Users)
}

// LoadUserByIDFromDir loads a given user from a given directory
func LoadUserByIDFromDir(baseDir string, id string) *User {
	return LoadUsersFromDir(baseDir).Users[id]
}

// LoadUsersFromDir loads all users from a given directory
func LoadUsersFromDir(baseDir string) GlobalUsers {
	files, _ := zglob.Glob(GetAptomiObjectDir(baseDir, Users) + "/**/users.*.yaml")

	r := GlobalUsers{Users: make(map[string]*User)}

	for _, f := range files {
		debug.WithFields(log.Fields{
			"file": f,
		}).Debug("Loading users")

		dat, e := ioutil.ReadFile(f)
		if e == nil {
			t := []*User{}
			e = yaml.Unmarshal([]byte(dat), &t)
			if e != nil {
				debug.WithFields(log.Fields{
					"file":  f,
					"error": e,
				}).Fatal("Unable to unmarshal users")
			}
			for _, u := range t {
				// load secrets
				u.Secrets = LoadUserSecretsByIDFromDir(baseDir, u.ID)

				// add user
				r.Users[u.ID] = u
			}
		}
	}

	return r
}
