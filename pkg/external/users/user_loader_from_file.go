package users

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"github.com/Aptomi/aptomi/pkg/util"
	"github.com/patrickmn/go-cache"
)

// UserLoaderFromFile allows aptomi to load users from a file
type UserLoaderFromFile struct {
	fileName             string
	cache                *cache.Cache
	domainAdminOverrides map[string]bool
}

// NewUserLoaderFromFile returns new UserLoaderFromFile
func NewUserLoaderFromFile(fileName string, domainAdminOverrides map[string]bool) UserLoader {
	return &UserLoaderFromFile{
		fileName:             fileName,
		cache:                cache.New(time.Minute, time.Minute),
		domainAdminOverrides: domainAdminOverrides,
	}
}

// LoadUsersAll loads all users
func (loader *UserLoaderFromFile) LoadUsersAll() *lang.GlobalUsers {
	// this can be called concurrently by the engine, so it needs to be thread safe
	cachedUsers, _ := loader.cache.Get("users")
	if cachedUsers != nil {
		return cachedUsers.(*lang.GlobalUsers)
	}

	// synchronize and retrieve users
	mutex := sync.Mutex{}
	mutex.Lock()
	defer func() { mutex.Unlock() }()

	result := &lang.GlobalUsers{Users: make(map[string]*lang.User)}
	userList := loadUsersFromFile(loader.fileName)
	for _, u := range userList {
		result.Users[strings.ToLower(u.Name)] = u
		if _, exist := loader.domainAdminOverrides[strings.ToLower(u.Name)]; exist {
			u.DomainAdmin = true
		}
	}
	loader.cache.Set("users", result, cache.DefaultExpiration)
	return result
}

// LoadUserByName loads a single user by name
func (loader *UserLoaderFromFile) LoadUserByName(name string) *lang.User {
	return loader.LoadUsersAll().Users[strings.ToLower(name)]
}

// Authenticate should authenticate a user by username/password.
// If user exists and username/password is correct, it should be returned.
// If a user doesn't exist or username/password is not correct, then nil should be returned together with an error.
func (loader *UserLoaderFromFile) Authenticate(name, password string) (*lang.User, error) {
	user := loader.LoadUserByName(name)
	if user == nil {
		return nil, fmt.Errorf("user '%s' does not exist", name)
	}

	if !util.ComparePasswords(user.PasswordHash, password) {
		return nil, fmt.Errorf("incorrect password")
	}

	return user, nil
}

// Summary returns summary as string
func (loader *UserLoaderFromFile) Summary() string {
	return strconv.Itoa(len(loader.LoadUsersAll().Users)) + " (from filesystem)"
}

// Loads users from file
func loadUsersFromFile(fileName string) []*lang.User {
	return *yaml.LoadObjectFromFileDefaultEmpty(fileName, &[]*lang.User{}).(*[]*lang.User)
}
