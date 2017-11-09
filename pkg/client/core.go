package client

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Core is the Core API client interface
type Core interface {
	Policy() Policy
	Endpoints() Endpoints
	Version() Version
}

// Policy is the interface for managing Policy
type Policy interface {
	Show() (*engine.PolicyData, error)
	Apply([]runtime.Object) (*api.PolicyUpdateResult, error)
	Delete([]string) (*api.PolicyUpdateResult, error)
}

// Endpoints is the interface for getting info about endpoints
type Endpoints interface {
	Show() (*api.Endpoints, error)
}

// Version is the interface for getting current server version
type Version interface {
	Show() (*api.Version, error)
}
