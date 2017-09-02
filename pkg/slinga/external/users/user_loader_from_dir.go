package users

import (
	"github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/language"
	"github.com/Aptomi/aptomi/pkg/slinga/language/yaml"
	"github.com/mattn/go-zglob"
	"strconv"
)

// UserLoaderFromDir allows aptomi to load users from files in a given directory
type UserLoaderFromDir struct {
	baseDir     string
	cachedUsers *language.GlobalUsers
}

// NewUserLoaderFromDir returns new UserLoaderFromDir, given a directory where files should be read from
func NewUserLoaderFromDir(baseDir string) UserLoader {
	return &UserLoaderFromDir{baseDir: baseDir}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromDir) LoadUsersAll() language.GlobalUsers {
	if loader.cachedUsers == nil {
		files, _ := zglob.Glob(db.GetAptomiObjectFilePatternYaml(loader.baseDir, db.TypeUsersFile))
		loader.cachedUsers = &language.GlobalUsers{Users: make(map[string]*language.User)}
		for _, fileName := range files {
			t := loadUsersFromFile(fileName)
			for _, u := range t {
				// load secrets
				u.Secrets = language.LoadUserSecretsByIDFromDir(loader.baseDir, u.ID)

				// add user
				loader.cachedUsers.Users[u.ID] = u
			}
		}
	}
	return *loader.cachedUsers
}

// LoadUserByID loads a single user by ID
func (loader *UserLoaderFromDir) LoadUserByID(id string) *language.User {
	return loader.LoadUsersAll().Users[id]
}

// Summary returns summary as string
func (loader *UserLoaderFromDir) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from filesystem)"
}

// Loads users from file
func loadUsersFromFile(fileName string) []*language.User {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*language.User{}).(*[]*language.User)
}
