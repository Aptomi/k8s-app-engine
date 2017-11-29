package users

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	"strconv"
	"strings"
)

// UserLoaderMock allows to mock user loader and use in-memory user storage.
// It also allows to emulate panics
type UserLoaderMock struct {
	users       *lang.GlobalUsers
	shouldPanic bool
}

// NewUserLoaderMock returns new UserLoaderMock
func NewUserLoaderMock() *UserLoaderMock {
	return &UserLoaderMock{
		users: &lang.GlobalUsers{Users: make(map[string]*lang.User)},
	}
}

// SetPanic controls whether user loader should panic or return users
func (loader *UserLoaderMock) SetPanic(panicFlag bool) {
	loader.shouldPanic = panicFlag
}

// AddUser adds a user into the mock structure
func (loader *UserLoaderMock) AddUser(user *lang.User) {
	loader.users.Users[strings.ToLower(user.Name)] = user
}

// LoadUsersAll loads all users
func (loader *UserLoaderMock) LoadUsersAll() *lang.GlobalUsers {
	if loader.shouldPanic {
		panic("panic from mock user loader")
	}
	return loader.users
}

// LoadUserByName loads a single user by Name
func (loader *UserLoaderMock) LoadUserByName(name string) *lang.User {
	if loader.shouldPanic {
		panic("panic from mock user loader")
	}
	return loader.users.Users[strings.ToLower(name)]
}

// Authenticate does nothing for mock
func (loader *UserLoaderMock) Authenticate(name, password string) (*lang.User, error) {
	if loader.shouldPanic {
		panic("panic from mock user loader")
	}
	return nil, nil
}

// Summary returns summary as string
func (loader *UserLoaderMock) Summary() string {
	return strconv.Itoa(len(loader.users.Users)) + " (mock)"
}
