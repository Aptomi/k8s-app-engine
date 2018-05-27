package client

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/version"
	"github.com/Sirupsen/logrus"
)

// Core is the Core API client interface
type Core interface {
	Policy() Policy
	Dependency() Dependency
	Revision() Revision
	State() State
	User() User
	Version() Version
}

// Policy is the interface for managing Policy
type Policy interface {
	Show(gen runtime.Generation) (*engine.PolicyData, error)
	Apply([]runtime.Object, bool, logrus.Level) (*api.PolicyUpdateResult, error)
	Delete([]runtime.Object, bool, logrus.Level) (*api.PolicyUpdateResult, error)
}

// Dependency is the interface for managing Dependency
type Dependency interface {
	Status([]*lang.Dependency, api.DependencyQueryFlag) (*api.DependenciesStatus, error)
}

// Revision is the interface for getting Revisions
type Revision interface {
	Show(gen runtime.Generation) (*engine.Revision, error)
}

// State is the interface for resetting Actual State
type State interface {
	Reset(bool) (*api.PolicyUpdateResult, error)
}

// User is the interface for auth and user management
type User interface {
	Login(username, password string) (*api.AuthSuccess, error)
}

// Version is the interface for getting current server version
type Version interface {
	Show() (*version.BuildInfo, error)
}
