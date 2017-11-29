package config

import (
	"sort"
)

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

// GetAttributes returns the list of attributes to be retrieved from LDAP
func (cfg *LDAP) GetAttributes() []string {
	result := []string{}
	for _, attr := range cfg.LabelToAttributes {
		result = append(result, attr)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
