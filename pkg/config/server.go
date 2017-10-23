package config

// Server represents configs for the server
type Server struct {
	Debug                bool
	API                  API
	DB                   DB
	Helm                 Helm
	Users                UserSources
	SecretsDir           string `valid:"dir"`
	Enforcer             Enforcer
	DomainAdminOverrides map[string]bool
}

// UserSources represents configs for the user loaders that could be file and LDAP loaders
type UserSources struct {
	LDAP []LDAP
	File []string
}

// IsDebug returns true if debug mode enabled
func (s Server) IsDebug() bool {
	return s.Debug
}

// LDAP contains configuration for LDAP sync service (host, port, DN, filter query and mapping of LDAP properties to Aptomi attributes)
type LDAP struct {
	Host              string
	Port              int
	BaseDN            string
	Filter            string
	LabelToAttributes map[string]string
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
	ChartsDir string `valid:"dir,required"`
}

// DB represents configs for DB
type DB struct {
	Connection string `valid:"required"`
}

// Enforcer represents configs for Enforcer background process that periodically gets latest policy, calculating
// difference between it and actual state and then applying calculated actions.
type Enforcer struct {
	Disabled bool
}
