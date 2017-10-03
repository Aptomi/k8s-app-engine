package users

import (
	"github.com/Aptomi/aptomi/pkg/slinga/db"
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"github.com/Aptomi/aptomi/pkg/slinga/lang/yaml"
	"github.com/mattn/go-zglob"
	"strconv"
	"sync"
)

// UserLoaderFromDir allows aptomi to load users from files in a given directory
type UserLoaderFromDir struct {
	once sync.Once

	baseDir     string
	cachedUsers *lang.GlobalUsers
}

// NewUserLoaderFromDir returns new UserLoaderFromDir, given a directory where files should be read from
func NewUserLoaderFromDir(baseDir string) UserLoader {
	return &UserLoaderFromDir{
		baseDir: baseDir,
	}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromDir) LoadUsersAll() lang.GlobalUsers {
	// Right now this can be called concurrently by the engine, so it needs to be thread safe
	loader.once.Do(func() {
		files, _ := zglob.Glob(db.GetAptomiObjectFilePatternYaml(loader.baseDir, db.TypeUsersFile))
		loader.cachedUsers = &lang.GlobalUsers{Users: make(map[string]*lang.User)}
		for _, fileName := range files {
			t := loadUsersFromFile(fileName)
			for _, u := range t {
				// add user
				loader.cachedUsers.Users[u.ID] = u
			}
		}
	})
	return *loader.cachedUsers
}

// LoadUserByID loads a single user by ID
func (loader *UserLoaderFromDir) LoadUserByID(id string) *lang.User {
	return loader.LoadUsersAll().Users[id]
}

// Summary returns summary as string
func (loader *UserLoaderFromDir) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from filesystem)"
}

// Loads users from file
func loadUsersFromFile(fileName string) []*lang.User {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*lang.User{}).(*[]*lang.User)
}
