package config

import "time"

// Server represents configs for the server
type Server struct {
	Debug                bool            `validate:"-"`
	API                  API             `validate:"required"`
	UI                   UI              `validate:"omitempty"` // if UI is not defined, then UI will not be started
	DB                   DB              `validate:"required"`
	Helm                 Helm            `validate:"required"`
	Users                UserSources     `validate:"required"`
	SecretsDir           string          `validate:"omitempty,dir"` // secrets is not a first-class citizen yet, so it's not required
	Enforcer             Enforcer        `validate:"required"`
	DomainAdminOverrides map[string]bool `validate:"-"`
	Secret               string          `validate:"required"`
}

// UserSources represents configs for the user loaders that could be file and LDAP loaders
type UserSources struct {
	LDAP []LDAP   `validate:"dive"`
	File []string `validate:"dive,file"`
}

// IsDebug returns true if debug mode enabled
func (s Server) IsDebug() bool {
	return s.Debug
}

// Helm represents configs for Helm plugin
type Helm struct {
}

// DB represents configs for DB
type DB struct {
	Connection string `validate:"required"`
}

// Enforcer represents configs for Enforcer background process that periodically gets latest policy, calculating
// difference between it and actual state and then applying calculated actions.
type Enforcer struct {
	Interval  time.Duration `validate:"-"`
	Disabled  bool          `validate:"-"`
	Noop      bool          `validate:"-"`
	NoopSleep int           `validate:"-"`
}
