package config

type Server struct {
	Debug      bool
	API        API
	DB         DB
	Helm       Helm
	Users      UserSources
	SecretsDir string `valid:"dir"`
	Enforcer   Enforcer
}

type UserSources struct {
	LDAP []LDAP
	File []string
}

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

type Helm struct {
	ChartsDir string `valid:"dir,required"`
}

type DB struct {
	Connection string `valid:"required"`
}

type Enforcer struct {
	Disabled bool
}
