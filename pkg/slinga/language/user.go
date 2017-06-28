package language

import (
	. "github.com/Frostman/aptomi/pkg/slinga/fileio"
	. "github.com/Frostman/aptomi/pkg/slinga/maputil"
	"github.com/mattn/go-zglob"
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

// Count returns total number of users
func (users *GlobalUsers) Count() int {
	return CountElements(users.Users)
}

// LoadUserByIDFromDir loads a given user from a given directory
func LoadUserByIDFromDir(baseDir string, id string) *User {
	return LoadUsersFromDir(baseDir).Users[id]
}

// LoadUsersFromDir loads all users from a given directory
func LoadUsersFromDir(baseDir string) GlobalUsers {
	files, _ := zglob.Glob(GetAptomiObjectFilePatternYaml(baseDir, TypeUsers))
	r := GlobalUsers{Users: make(map[string]*User)}
	for _, fileName := range files {
		t := loadUsersFromFile(fileName)
		for _, u := range t {
			// load secrets
			u.Secrets = LoadUserSecretsByIDFromDir(baseDir, u.ID)

			// add user
			r.Users[u.ID] = u
		}
	}

	return r
}

// LoadUsers loads users from current users.yaml
func LoadUsers() GlobalUsers {
	return LoadUsersFromDir(GetAptomiBaseDir())
}

// IsGlobalOps returns if user is a global ops guy
func (user *User) IsGlobalOps() bool {
	return user.Labels["global_ops"] == "true"
}
