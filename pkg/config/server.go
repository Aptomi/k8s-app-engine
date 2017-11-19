package config

// Server represents configs for the server
type Server struct {
	Debug                bool            `validate:"-"`
	API                  API             `validate:"required"`
	DB                   DB              `validate:"required"`
	Helm                 Helm            `validate:"required"`
	Users                UserSources     `validate:"required"`
	SecretsDir           string          `validate:"omitempty,dir"` // secrets is not a first-class citizen yet, so it's not required
	Enforcer             Enforcer        `validate:"required"`
	DomainAdminOverrides map[string]bool `validate:"-"`
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

// LDAP contains configuration for LDAP sync service (host, port, DN, filter query and mapping of LDAP properties to Aptomi attributes)
type LDAP struct {
	Host   string `validate:"required,hostname|ip"`
	Port   int    `validate:"required,min=1,max=65535"`
	BaseDN string `validate:"required"`

	// Filter is LDAP filter query for all users
	Filter string `validate:"required"`

	// FilterByName is LDAP filter query when doing user lookup by name
	FilterByName string `validate:"required"`

	LabelToAttributes map[string]string `validate:"required"`
}

// GetAttributes Returns the list of attributes to be retrieved from LDAP
func (cfg *LDAP) GetAttributes() []string {
	result := []string{}
	for _, attr := range cfg.LabelToAttributes {
		result = append(result, attr)
	}
	return result
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
	Disabled  bool `validate:"-"`
	Noop      bool `validate:"-"`
	NoopSleep int  `validate:"-"`
}
