package config

import (
	"github.com/Sirupsen/logrus"
	"time"
)

// Server represents configs for the server
type Server struct {
	Debug                bool                 `validate:"-"`
	API                  API                  `validate:"required"`
	UI                   UI                   `validate:"omitempty"` // if UI is not defined, then UI will not be started
	DB                   DB                   `validate:"required"`
	Plugins              Plugins              `validate:"required"`
	Users                UserSources          `validate:"required"`
	SecretsDir           string               `validate:"omitempty,dir"` // secrets is not a first-class citizen yet, so it's not required
	Enforcer             DesiredStateEnforcer `validate:"required"`
	Updater              ActualStateUpdater   `validate:"required"`
	DomainAdminOverrides map[string]bool      `validate:"-"`
	Auth                 ServerAuth           `validate:"-"`
	Profile              Profile              `validate:"-"`
}

// IsDebug returns true if debug mode enabled
func (s Server) IsDebug() bool {
	return s.Debug
}

// GetLogLevel returns log level
func (s *Server) GetLogLevel() logrus.Level {
	if s.IsDebug() {
		return logrus.DebugLevel
	}
	return logrus.InfoLevel
}

// UserSources represents configs for the user loaders that could be file and LDAP loaders
type UserSources struct {
	LDAP []LDAP   `validate:"dive"`
	File []string `validate:"dive,file"`
}

// DB represents configs for DB
type DB struct {
	Connection string `validate:"required"`
}

// DesiredStateEnforcer represents config for desired state enforcer background process that periodically gets latest policy, calculating
// difference between it and actual state and then applying calculated actions
type DesiredStateEnforcer struct {
	Disabled             bool          `validate:"-"`
	Interval             time.Duration `validate:"-"`
	Noop                 bool          `validate:"-"`
	NoopSleep            time.Duration `validate:"-"`
	MaxConcurrentActions int           `validate:"-"`
}

// ActualStateUpdater represents config for actual state updater background process that periodically refreshes actual state
// (e.g. retrieves endpoints for all components)
type ActualStateUpdater struct {
	Disabled             bool          `validate:"-"`
	Interval             time.Duration `validate:"-"`
	Noop                 bool          `validate:"-"`
	NoopSleep            time.Duration `validate:"-"`
	MaxConcurrentActions int           `validate:"-"`
}

// ServerAuth represents server auth config
type ServerAuth struct {
	Secret string `validate:"-"`
}

// Profile represents profiler config
type Profile struct {
	CPU   string
	Trace string
}
