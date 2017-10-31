package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"strconv"
	"strings"
	"sync"
)

// UserLoaderFromFile allows aptomi to load users from a file
type UserLoaderFromFile struct {
	once sync.Once

	fileName             string
	users                *lang.GlobalUsers
	domainAdminOverrides map[string]bool
}

// NewUserLoaderFromFile returns new UserLoaderFromFile
func NewUserLoaderFromFile(fileName string, domainAdminOverrides map[string]bool) UserLoader {
	return &UserLoaderFromFile{
		fileName:             fileName,
		domainAdminOverrides: domainAdminOverrides,
	}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromFile) LoadUsersAll() *lang.GlobalUsers {
	// Right now this can be called concurrently by the engine, so it needs to be thread safe
	loader.once.Do(func() {
		loader.users = &lang.GlobalUsers{Users: make(map[string]*lang.User)}
		t := loadUsersFromFile(loader.fileName)
		for _, u := range t {
			loader.users.Users[u.Name] = u
			if _, exist := loader.domainAdminOverrides[strings.ToLower(u.Name)]; exist {
				u.Admin = true
			}
		}
	})
	return loader.users
}

// LoadUserByName loads a single user by name
func (loader *UserLoaderFromFile) LoadUserByName(name string) *lang.User {
	return loader.LoadUsersAll().Users[name]
}

// Summary returns summary as string
func (loader *UserLoaderFromFile) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from filesystem)"
}

// Loads users from file
func loadUsersFromFile(fileName string) []*lang.User {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*lang.User{}).(*[]*lang.User)
}
