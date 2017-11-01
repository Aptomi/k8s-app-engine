package users

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/lang/yaml"
	"golang.org/x/crypto/bcrypt"
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
			u.Name = strings.ToLower(u.Name)
			loader.users.Users[u.Name] = u
			if _, exist := loader.domainAdminOverrides[u.Name]; exist {
				u.Admin = true
			}
		}
	})
	return loader.users
}

// LoadUserByName loads a single user by name
func (loader *UserLoaderFromFile) LoadUserByName(name string) *lang.User {
	name = strings.ToLower(name)
	return loader.LoadUsersAll().Users[name]
}

// Authenticate should authenticate a user by username/password.
// If user exists and username/password is correct, it should be returned.
// If a user doesn't exist or username/password is not correct, then nil should be returned together with an error.
func (loader *UserLoaderFromFile) Authenticate(name, password string) (*lang.User, error) {
	user := loader.LoadUserByName(name)
	if user == nil {
		return nil, fmt.Errorf("user '%s' does not exist", name)
	}

	if !comparePasswords(user.PasswordHash, password) {
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

// Returns salted hash from the password (only used to generate user passwords)
func hashAndSalt(password string) string { // nolint: deadcode, megacheck
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// Verifies hashed password
func comparePasswords(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
