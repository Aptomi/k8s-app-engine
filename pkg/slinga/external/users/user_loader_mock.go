package users

import (
	"github.com/Aptomi/aptomi/pkg/slinga/lang"
	"strconv"
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
	loader.users.Users[user.ID] = user
}

// LoadUsersAll loads all users
func (loader *UserLoaderMock) LoadUsersAll() *lang.GlobalUsers {
	return loader.users
}

// LoadUserByID loads a single user by ID
func (loader *UserLoaderMock) LoadUserByID(id string) *lang.User {
	return loader.users.Users[id]
}

// Summary returns summary as string
func (loader *UserLoaderMock) Summary() string {
	return strconv.Itoa(len(loader.users.Users)) + " (mock)"
}
