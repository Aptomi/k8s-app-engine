package client

import (
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

type Core interface {
	Policy() Policy
	Endpoints() Endpoints
	Version() Version
}

type Policy interface {
	Show() (*engine.PolicyData, error)
	Apply([]runtime.Object) (*api.PolicyUpdateResult, error)
	Delete([]string) (*api.PolicyUpdateResult, error)
}

type Endpoints interface {
	Show() (*api.Endpoints, error)
}

type Version interface {
	Show() (*api.Version, error)
}
