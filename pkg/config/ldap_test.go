package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigLDAP(t *testing.T) {
	config := &LDAP{
		LabelToAttributes: map[string]string{
			"aptomi_label_1": "ldap_attr_1",
			"aptomi_label_2": "ldap_attr_2",
			"aptomi_label_3": "ldap_attr_3",
		},
	}
	attrNames := config.GetAttributes()
	assert.Equal(t, []string{"ldap_attr_1", "ldap_attr_2", "ldap_attr_3"}, attrNames, "The list of attributes to be retrieved from LDAP must be correct")
}
