package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"strconv"
	"strings"
)

// UserLoaderMock allows to mock user loader and use in-memory user storage
type UserLoaderMock struct {
	users *lang.GlobalUsers
}

// NewUserLoaderMock returns new UserLoaderMock
func NewUserLoaderMock() *UserLoaderMock {
	return &UserLoaderMock{
		users: &lang.GlobalUsers{Users: make(map[string]*lang.User)},
	}
}

// AddUser adds a user into the mock structure
func (loader *UserLoaderMock) AddUser(user *lang.User) {
	user.Name = strings.ToLower(user.Name)
	loader.users.Users[user.Name] = user
}

// LoadUsersAll loads all users
func (loader *UserLoaderMock) LoadUsersAll() *lang.GlobalUsers {
	return loader.users
}

// LoadUserByName loads a single user by Name
func (loader *UserLoaderMock) LoadUserByName(name string) *lang.User {
	return loader.users.Users[strings.ToLower(name)]
}

// Authenticate does nothing for mock
func (loader *UserLoaderMock) Authenticate(name, password string) (*lang.User, error) {
	return nil, nil
}

// Summary returns summary as string
func (loader *UserLoaderMock) Summary() string {
	return strconv.Itoa(len(loader.users.Users)) + " (mock)"
}
